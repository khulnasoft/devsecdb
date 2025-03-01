<template>
  <div class="focus:outline-none" tabindex="0" v-bind="$attrs">
    <NoPermissionPlaceholder v-if="!hasPermission" />
    <main v-else-if="changeHistory" class="flex flex-col relative gap-y-6">
      <!-- Highlight Panel -->
      <div
        class="pb-4 border-b border-block-border md:flex md:items-center md:justify-between"
      >
        <div class="flex-1 min-w-0 space-y-3">
          <!-- Summary -->
          <div class="flex items-center space-x-2">
            <ChangeHistoryStatusIcon :status="changeHistory.status" />
            <h1 class="text-xl font-bold leading-6 text-main truncate">
              {{ $t("common.version") }} {{ changeHistory.version }}
            </h1>
            <NTag round>
              {{ changeHistory_TypeToJSON(changeHistory.type) }}
            </NTag>
          </div>
          <dl
            class="flex flex-col space-y-1 md:space-y-0 md:flex-row md:flex-wrap"
          >
            <dt class="sr-only">{{ $t("common.issue") }}</dt>
            <dd class="flex items-center text-sm md:mr-4">
              <span class="textlabel"
                >{{ $t("common.issue") }}&nbsp;-&nbsp;</span
              >
              <router-link
                :to="{
                  path: `/${changeHistory.issue}`,
                }"
                class="normal-link"
              >
                #{{ extractIssueUID(changeHistory.issue) }}
              </router-link>
            </dd>
            <dt class="sr-only">{{ $t("common.creator") }}</dt>
            <dd v-if="creator" class="flex items-center text-sm md:mr-4">
              <span class="textlabel"
                >{{ $t("common.creator") }}&nbsp;-&nbsp;</span
              >
              {{ creator.title }}
            </dd>
            <dt class="sr-only">{{ $t("common.created-at") }}</dt>
            <dd class="flex items-center text-sm md:mr-4">
              <span class="textlabel"
                >{{ $t("common.created-at") }}&nbsp;-&nbsp;</span
              >
              {{
                humanizeDate(getDateForPbTimestamp(changeHistory.createTime))
              }}
            </dd>
          </dl>
        </div>
      </div>

      <div v-if="affectedTables.length > 0">
        <span class="flex items-center text-lg text-main capitalize">
          {{ $t("change-history.affected-tables") }}
        </span>
        <div>
          <span
            v-for="(affectedTable, i) in affectedTables"
            :key="`${i}.${affectedTable.schema}.${affectedTable.table}`"
            :class="[
              'mr-3 mb-2',
              !affectedTable.dropped
                ? 'text-blue-600 cursor-pointer hover:opacity-80'
                : 'mb-2 text-gray-400 italic',
            ]"
            @click="handleAffectedTableClick(affectedTable)"
            >{{ getAffectedTableDisplayName(affectedTable) }}</span
          >
        </div>
      </div>

      <div class="flex flex-col gap-y-6">
        <div class="flex flex-col gap-y-2">
          <a
            id="statement"
            href="#statement"
            class="w-auto flex items-center text-lg text-main mb-2 hover:underline"
          >
            {{ $t("common.statement") }}
            <button
              tabindex="-1"
              class="btn-icon ml-1"
              @click.prevent="copyStatement"
            >
              <heroicons-outline:clipboard class="w-6 h-6" />
            </button>
          </a>
          <MonacoEditor
            class="h-auto max-h-[480px] min-h-[120px] border rounded-[3px] text-sm overflow-clip relative"
            :content="changeHistoryStatement"
            :readonly="true"
            :auto-height="{ min: 120, max: 480 }"
          />
          <div v-if="guessedIsBasicView">
            <NButton
              quaternary
              size="small"
              :disabled="state.loading"
              @click="fetchFullHistory"
            >
              <template #icon>
                <BBSpin v-if="state.loading" />
                <ChevronDownIcon v-else class="w-5" />
              </template>
              {{ $t("change-history.view-full") }}
            </NButton>
          </div>
        </div>
        <div v-if="showSchemaSnapshot" class="flex flex-col gap-y-2">
          <a
            id="schema"
            href="#schema"
            class="flex items-center text-lg text-main hover:underline capitalize"
          >
            Schema {{ $t("common.snapshot") }}
            <button
              tabindex="-1"
              class="btn-icon ml-1"
              @click.prevent="copySchema"
            >
              <heroicons-outline:clipboard class="w-6 h-6" />
            </button>
          </a>

          <div v-if="hasDrift" class="flex items-center gap-x-2">
            <div class="flex items-center text-sm font-normal">
              <heroicons-outline:exclamation-circle
                class="w-5 h-5 mr-0.5 text-error"
              />
              <span>{{ $t("change-history.schema-drift-detected") }}</span>
            </div>
            <div
              class="normal-link text-sm"
              data-label="bb-change-history-view-drift-button"
              @click="state.viewDrift = true"
            >
              {{ $t("change-history.view-drift") }}
            </div>
          </div>

          <div class="flex flex-row items-center gap-x-2">
            <div v-if="allowShowDiff" class="flex space-x-1 items-center">
              <NSwitch
                :value="state.showDiff"
                size="small"
                :disabled="state.loading"
                data-label="bb-change-history-diff-switch"
                @update:value="switchShowDiff"
              />
              <span class="text-sm font-semibold">
                {{ $t("change-history.show-diff") }}
              </span>
            </div>
            <div class="textinfolabel">
              <i18n-t
                v-if="state.showDiff"
                tag="span"
                keypath="change-history.left-vs-right"
              >
                <template #prevLink>
                  <router-link
                    v-if="previousHistory"
                    class="normal-link"
                    :to="previousHistoryLink"
                  >
                    ({{ previousHistory.version }})
                  </router-link>
                </template>
              </i18n-t>
              <template v-else>
                {{ $t("change-history.schema-snapshot-after-change") }}
              </template>
            </div>
            <div v-if="!allowShowDiff" class="text-sm font-normal text-accent">
              ({{ $t("change-history.no-schema-change") }})
            </div>
          </div>

          <DiffEditor
            v-if="state.showDiff"
            class="h-auto max-h-[600px] min-h-[120px] border rounded-md text-sm overflow-clip"
            :original="changeHistory.prevSchema"
            :modified="changeHistory.schema"
            :readonly="true"
            :auto-height="{ min: 120, max: 600 }"
          />
          <template v-else>
            <div v-if="changeHistory.schema" class="space-y-2">
              <MonacoEditor
                class="h-auto max-h-[600px] min-h-[120px] border rounded-md text-sm overflow-clip relative"
                :content="changeHistorySchema"
                :readonly="true"
                :auto-height="{ min: 120, max: 600 }"
              />
              <div
                v-if="
                  getStatementSize(changeHistory.schema).ne(
                    changeHistory.schemaSize
                  )
                "
              >
                <NButton
                  quaternary
                  size="small"
                  :disabled="state.loading"
                  @click="fetchFullHistory"
                >
                  <template #icon>
                    <BBSpin v-if="state.loading" />
                    <ChevronDownIcon v-else class="w-5" />
                  </template>
                  {{ $t("change-history.view-full") }}
                </NButton>
              </div>
            </div>
            <div v-else>
              {{ $t("change-history.current-schema-empty") }}
            </div>
          </template>
        </div>
      </div>
    </main>

    <BBModal
      v-if="changeHistory && previousHistory && state.viewDrift"
      @close="state.viewDrift = false"
    >
      <template #title>
        <span>{{ $t("change-history.schema-drift") }}</span>
        <span class="mx-2">-</span>
        <i18n-t tag="span" keypath="change-history.left-vs-right">
          <template #prevLink>
            <router-link class="normal-link" :to="previousHistoryLink">
              ({{ previousHistory.version }})
            </router-link>
          </template>
        </i18n-t>
      </template>

      <div
        class="space-y-4 flex flex-col overflow-hidden"
        style="width: calc(100vw - 10rem); height: calc(100vh - 12rem)"
      >
        <DiffEditor
          class="flex-1 w-full border rounded-md overflow-clip"
          :original="previousHistory.schema"
          :modified="changeHistory.schema"
          :readonly="true"
        />
        <div class="flex justify-end">
          <NButton type="primary" @click.prevent="state.viewDrift = false">
            {{ $t("common.close") }}
          </NButton>
        </div>
      </div>
    </BBModal>
  </div>

  <TableDetailDrawer
    :show="!!selectedAffectedTable"
    :database-name="database.name"
    :schema-name="selectedAffectedTable?.schema ?? ''"
    :table-name="selectedAffectedTable?.table ?? ''"
    :classification-config="classificationConfig"
    @dismiss="selectedAffectedTable = undefined"
  />
</template>

<script lang="ts" setup>
import { useTitle } from "@vueuse/core";
import { ChevronDownIcon } from "lucide-vue-next";
import { NButton, NSwitch, NTag } from "naive-ui";
import { computed, reactive, watch, ref, unref } from "vue";
import { BBModal, BBSpin } from "@/bbkit";
import ChangeHistoryStatusIcon from "@/components/ChangeHistory/ChangeHistoryStatusIcon.vue";
import { DiffEditor, MonacoEditor } from "@/components/MonacoEditor";
import TableDetailDrawer from "@/components/TableDetailDrawer.vue";
import {
  pushNotification,
  useChangeHistoryStore,
  useDBSchemaV1Store,
  useUserStore,
  useSettingV1Store,
  useDatabaseV1ByName,
} from "@/store";
import { getDateForPbTimestamp } from "@/types";
import type { AffectedTable } from "@/types/changeHistory";
import { Engine } from "@/types/proto/v1/common";
import type { ChangeHistory } from "@/types/proto/v1/database_service";
import {
  ChangeHistory_Type,
  changeHistory_TypeToJSON,
  ChangeHistoryView,
} from "@/types/proto/v1/database_service";
import {
  changeHistoryLink,
  extractIssueUID,
  extractUserResourceName,
  getAffectedTablesOfChangeHistory,
  toClipboard,
  getStatementSize,
  hasProjectPermissionV2,
  getAffectedTableDisplayName,
  extractChangeHistoryUID,
} from "@/utils";
import NoPermissionPlaceholder from "../misc/NoPermissionPlaceholder.vue";

interface LocalState {
  showDiff: boolean;
  viewDrift: boolean;
  loading: boolean;
}

const props = defineProps<{
  instance: string;
  database: string;
  changeHistoryId: string;
}>();

const state = reactive<LocalState>({
  showDiff: false,
  viewDrift: false,
  loading: false,
});

const dbSchemaStore = useDBSchemaV1Store();
const settingStore = useSettingV1Store();
const changeHistoryStore = useChangeHistoryStore();
const selectedAffectedTable = ref<AffectedTable | undefined>();

const { database } = useDatabaseV1ByName(props.database);

const hasPermission = computed(() =>
  hasProjectPermissionV2(database.value.projectEntity, "bb.changeHistories.get")
);

const classificationConfig = computed(() => {
  return settingStore.getProjectClassification(
    database.value.projectEntity.dataClassificationConfigId
  );
});

const changeHistoryName = computed(() => {
  return `${props.database}/changeHistories/${props.changeHistoryId}`;
});

const affectedTables = computed(() => {
  if (changeHistory.value === undefined) {
    return [];
  }
  return getAffectedTablesOfChangeHistory(changeHistory.value);
});

const showSchemaSnapshot = computed(() => {
  return database.value.instanceResource.engine !== Engine.RISINGWAVE;
});

watch(
  [database.value.name, changeHistoryName],
  async ([_, name]) => {
    await Promise.all([
      dbSchemaStore.getOrFetchDatabaseMetadata({
        database: database.value.name,
        skipCache: false,
      }),
      changeHistoryStore.getOrFetchChangeHistoryByName(
        unref(name),
        ChangeHistoryView.CHANGE_HISTORY_VIEW_FULL
      ),
    ]);
  },
  { immediate: true }
);

const switchShowDiff = async (showDiff: boolean) => {
  await fetchFullHistory();
  state.showDiff = showDiff;
};

const handleAffectedTableClick = (affectedTable: AffectedTable): void => {
  if (affectedTable.dropped) {
    return;
  }
  selectedAffectedTable.value = affectedTable;
};

// get all change histories before (include) the one of given id, ordered by descending version.
const prevChangeHistoryList = computed(() => {
  const changeHistoryList = changeHistoryStore.changeHistoryListByDatabase(
    database.value.name
  );

  // The returned change history list has been ordered by `id` DESC or (`namespace` ASC, `sequence` DESC) .
  // We can obtain prevChangeHistoryList by cutting up the array by the `changeHistoryId`.
  const idx = changeHistoryList.findIndex(
    (history) => extractChangeHistoryUID(history.name) === props.changeHistoryId
  );
  if (idx === -1) {
    return [];
  }
  return changeHistoryList.slice(idx);
});

// changeHistory is the latest migration NOW.
const changeHistory = computed((): ChangeHistory | undefined => {
  if (prevChangeHistoryList.value.length > 0) {
    const current = prevChangeHistoryList.value[0];
    return changeHistoryStore.getChangeHistoryByName(current.name) ?? current;
  }
  return changeHistoryStore.getChangeHistoryByName(changeHistoryName.value);
});

const changeHistorySchema = computed(() => {
  if (!changeHistory.value) {
    return "";
  }
  let schema = changeHistory.value.schema;
  if (guessedIsBasicView.value) {
    schema = `${schema}${schema.endsWith("\n") ? "" : "\n"}...`;
  }
  return schema;
});

const changeHistoryStatement = computed(() => {
  if (!changeHistory.value) {
    return "";
  }
  let statement = changeHistory.value.statement;
  if (
    getStatementSize(changeHistory.value.statement).lt(
      changeHistory.value.statementSize
    )
  ) {
    statement = `${statement}${statement.endsWith("\n") ? "" : "\n"}...`;
  }
  return statement;
});

// previousHistory is the last change history before the one of given id.
// Only referenced if hasDrift is true.
const previousHistory = computed((): ChangeHistory | undefined => {
  const prev = prevChangeHistoryList.value[1];
  if (!prev) return undefined;
  return changeHistoryStore.getChangeHistoryByName(prev.name) ?? prev;
});

const fetchFullPreviousHistory = async () => {
  const prev = previousHistory.value;
  if (!prev) return;
  await changeHistoryStore.getOrFetchChangeHistoryByName(
    prev.name,
    ChangeHistoryView.CHANGE_HISTORY_VIEW_FULL
  );
};

const fetchFullHistory = async () => {
  if (state.loading) {
    return;
  }
  state.loading = true;
  try {
    await Promise.all([
      changeHistoryStore.getOrFetchChangeHistoryByName(
        changeHistoryName.value,
        ChangeHistoryView.CHANGE_HISTORY_VIEW_FULL
      ),
      fetchFullPreviousHistory(),
    ]);
  } finally {
    state.loading = false;
  }
};

const guessedIsBasicView = computed(() => {
  const history = changeHistory.value;
  if (!history) return true;
  return getStatementSize(history.statement).ne(history.statementSize);
});

// "Show diff" feature is enabled when current migration has changed the schema.
const allowShowDiff = computed((): boolean => {
  if (!changeHistory.value) {
    return false;
  }
  return true;
});

// A schema drift is detected when the schema AFTER previousHistory has been
// changed unexpectedly BEFORE current changeHistory.
const hasDrift = computed((): boolean => {
  if (!changeHistory.value) {
    return false;
  }
  if (changeHistory.value.type === ChangeHistory_Type.BASELINE) {
    return false;
  }
  if (guessedIsBasicView.value) {
    return false;
  }

  return (
    prevChangeHistoryList.value.length > 1 && // no drift if no previous change history
    previousHistory.value?.schema !== changeHistory.value.prevSchema
  );
});

const creator = computed(() => {
  if (!changeHistory.value) {
    return undefined;
  }
  const email = extractUserResourceName(changeHistory.value.creator);
  return useUserStore().getUserByEmail(email);
});

const previousHistoryLink = computed(() => {
  const previous = previousHistory.value;
  if (!previous) return "";
  return changeHistoryLink(previous);
});

const copyStatement = async () => {
  await fetchFullHistory();

  if (!changeHistoryStatement.value) {
    return false;
  }
  toClipboard(changeHistoryStatement.value).then(() => {
    pushNotification({
      module: "devsecdb",
      style: "INFO",
      title: `Statement copied to clipboard.`,
    });
  });
};

const copySchema = async () => {
  await fetchFullHistory();

  if (!changeHistorySchema.value) {
    return false;
  }
  toClipboard(changeHistorySchema.value).then(() => {
    pushNotification({
      module: "devsecdb",
      style: "INFO",
      title: `Schema copied to clipboard.`,
    });
  });
};

watch(
  guessedIsBasicView,
  (basic) => {
    if (!basic) {
      fetchFullPreviousHistory();
    }
  },
  {
    immediate: true,
  }
);

useTitle(changeHistory.value?.version || "Change History");
</script>
