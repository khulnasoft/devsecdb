<template>
  <div
    class="w-full h-full flex flex-col gap-y-3 relative overflow-y-hidden overflow-x-auto pt-0.5"
  >
    <div
      class="w-[32rem] grid gap-y-3 gap-x-4 whitespace-nowrap items-center"
      style="grid-template-columns: minmax(auto, 8rem) 1fr"
    >
      <div class="contents">
        <div class="text-sm">
          {{ $t("database.branch-name") }}
        </div>
        <BBTextField
          v-model:value="branchId"
          required
          class="!w-60 text-sm"
          :placeholder="'feature/add-billing'"
        />
      </div>
      <div class="contents">
        <div class="text-sm">
          {{ $t("branch.source.self") }}
        </div>
        <NRadioGroup
          :value="source"
          class="!flex flex-row gap-x-2"
          @update:value="handleSwitchSource"
        >
          <NRadio value="PARENT">
            {{ $t("branch.source.parent-branch") }}
          </NRadio>
          <NTooltip :disabled="allowCreateBranchFromDatabase">
            <template #trigger>
              <NRadio
                :disabled="!allowCreateBranchFromDatabase"
                value="BASELINE"
              >
                {{ $t("branch.source.baseline-version") }}
              </NRadio>
            </template>
            <template #default>
              <div class="whitespace-nowrap">
                {{ $t("common.permission-denied") }}
              </div>
            </template>
          </NTooltip>
        </NRadioGroup>
      </div>
      <div v-if="source === 'PARENT'" class="contents">
        <div class="text-sm">
          {{ $t("schema-designer.parent-branch") }}
        </div>
        <BranchSelector
          v-model:branch="parentBranchName"
          :project="project"
          :loading="isPreparingBranch"
          :filter="filterParentBranch"
          class=""
          clearable
        />
      </div>
      <BaselineSchemaSelector
        v-if="source === 'BASELINE'"
        v-model:database-name="databaseName"
        :project-name="project.name"
        :loading="isPreparingBranch"
      />
    </div>
    <NDivider class="!my-0" />
    <div class="w-full flex-1 overflow-y-hidden">
      <SchemaEditorLite
        :key="branchData?.branch.name ?? ''"
        :loading="isPreparingBranch"
        :project="project"
        :resource-type="'branch'"
        :branch="branchData?.branch ?? EMPTY_BRANCH"
        :readonly="true"
        :diff-when-ready="!!branchData?.parent"
      />
      <!-- turn on diff-when-ready is useful when the branch is created from a parent branch -->
    </div>
    <div class="w-full flex items-center justify-end">
      <NButton
        type="primary"
        :disabled="!allowConfirm"
        :loading="isCreating"
        @click.prevent="handleConfirm"
      >
        {{ confirmText }}
      </NButton>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useDebounce } from "@vueuse/core";
import { cloneDeep, uniqueId } from "lodash-es";
import { NButton, NDivider, NRadio, NRadioGroup, NTooltip } from "naive-ui";
import { computed, ref, shallowRef, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { BBTextField } from "@/bbkit";
import SchemaEditorLite from "@/components/SchemaEditorLite";
import { PROJECT_V1_ROUTE_BRANCH_DETAIL } from "@/router/dashboard/projectV1";
import {
  pushNotification,
  useDatabaseV1Store,
  useDBSchemaV1Store,
} from "@/store";
import { useBranchStore } from "@/store/modules/branch";
import type { ComposedProject } from "@/types";
import { isValidDatabaseName } from "@/types";
import { Branch } from "@/types/proto/v1/branch_service";
import { DatabaseMetadataView } from "@/types/proto/v1/database_service";
import { hasProjectPermissionV2 } from "@/utils";
import BaselineSchemaSelector from "./BaselineSchemaSelector.vue";
import BranchSelector from "./BranchSelector.vue";
import { validateBranchName } from "./utils";

type Source = "PARENT" | "BASELINE";

type BranchData = {
  branch: Branch;
  parent: string | undefined;
};

const props = defineProps<{
  project: ComposedProject;
}>();

const DEBOUNCE_RATE = 100;
const { t } = useI18n();
const router = useRouter();
const databaseStore = useDatabaseV1Store();
const branchStore = useBranchStore();
const dbSchemaStore = useDBSchemaV1Store();
const source = ref<Source>("PARENT");
const databaseName = ref<string>();
const parentBranchName = ref<string>();
const isCreating = ref(false);
const branchId = ref<string>("");
const isPreparingBranch = ref(false);

const EMPTY_BRANCH = Branch.fromPartial({});

const allowCreateBranchFromDatabase = computed(() => {
  return hasProjectPermissionV2(props.project, "bb.branches.admin");
});

const debouncedDatabaseName = useDebounce(databaseName, DEBOUNCE_RATE);
const debouncedParentBranchName = useDebounce(parentBranchName, DEBOUNCE_RATE);

const filterParentBranch = (branch: Branch) => {
  // Only "main branch" aka parent-less branches can be parents.
  return !branch.parentBranch;
};

const nextFakeBranchName = () => {
  return `${props.project.name}/branches/-${uniqueId()}`;
};

const prepareBranchFromParentBranch = async (parent: string) => {
  const tag = `prepareBranchFromParentBranch(${parent})`;
  console.time(tag);
  const parentBranch = await branchStore.fetchBranchByName(
    parent,
    false /* !useCache */
  );
  const branch = cloneDeep(parentBranch);
  branch.name = nextFakeBranchName();
  console.timeEnd(tag);
  return branch;
};
const prepareBranchFromDatabaseHead = async (databaseName: string) => {
  const tag = `prepareBranchFromDatabaseHead(${databaseName})`;
  console.time(tag);

  console.time("--fetch metadata");
  const database = databaseStore.getDatabaseByName(databaseName);
  const metadata = await dbSchemaStore.getOrFetchDatabaseMetadata({
    database: database.name,
    view: DatabaseMetadataView.DATABASE_METADATA_VIEW_FULL,
    skipCache: true,
  });
  console.timeEnd("--fetch metadata");

  console.time("--build branch object");
  // Here metadata is not used for editing, so we need not to clone a copy
  // for baseline
  const branch = Branch.fromPartial({
    name: nextFakeBranchName(),
    engine: database.instanceResource.engine,
    baselineDatabase: database.name,
    baselineSchemaMetadata: metadata,
    schemaMetadata: metadata,
  });
  console.timeEnd("--build branch object");

  console.timeEnd(tag);
  return branch;
};

const branchData = shallowRef<BranchData>();

const prepareBranch = async (
  _parentBranchName: string | undefined,
  _databaseName: string | undefined
) => {
  isPreparingBranch.value = true;

  const finish = (s: BranchData | undefined) => {
    const isOutdated =
      _parentBranchName !== parentBranchName.value ||
      _databaseName !== databaseName.value;
    if (isOutdated) {
      return;
    }

    branchData.value = s;
    isPreparingBranch.value = false;
  };

  if (_parentBranchName) {
    const branch = await prepareBranchFromParentBranch(_parentBranchName);
    return finish({
      branch,
      parent: _parentBranchName,
    });
  }
  if (isValidDatabaseName(_databaseName)) {
    const branch = await prepareBranchFromDatabaseHead(_databaseName);
    return finish({
      branch,
      parent: undefined,
    });
  }
  return finish(undefined);
};

const handleSwitchSource = (src: Source) => {
  source.value = src;
  if (src === "PARENT") {
    databaseName.value = undefined;
  } else {
    parentBranchName.value = undefined;
  }
};

watch(
  [debouncedParentBranchName, debouncedDatabaseName],
  ([parentBranchName, databaseName]) => {
    prepareBranch(parentBranchName, databaseName);
  }
);

const allowConfirm = computed(() => {
  if (!hasProjectPermissionV2(props.project, "bb.branches.create")) {
    return false;
  }
  return branchId.value && branchData.value && !isCreating.value;
});

const confirmText = computed(() => {
  return t("common.create");
});

const handleConfirm = async () => {
  if (!branchData.value) {
    return;
  }

  if (!validateBranchName(branchId.value)) {
    pushNotification({
      module: "devsecdb",
      style: "CRITICAL",
      title: "Branch name valid characters: /^[a-zA-Z][a-zA-Z0-9-_/]+$/",
    });
    return;
  }

  const { branch, parent } = branchData.value;

  const { baselineDatabase } = branch;

  try {
    isCreating.value = true;
    if (!parent) {
      await branchStore.createBranch(
        props.project.name,
        branchId.value,
        Branch.fromPartial({
          baselineDatabase,
        })
      );
    } else {
      await branchStore.createBranch(
        props.project.name,
        branchId.value,
        Branch.fromPartial({
          parentBranch: parent,
        })
      );
    }
    isCreating.value = false;
    pushNotification({
      module: "devsecdb",
      style: "SUCCESS",
      title: t("schema-designer.message.created-succeed"),
    });

    // Go to branch detail page after created.
    router.replace({
      name: PROJECT_V1_ROUTE_BRANCH_DETAIL,
      params: {
        branchName: branchId.value,
      },
    });
  } catch {
    isCreating.value = false;
  }
};
</script>
