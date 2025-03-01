<template>
  <div class="space-y-4" v-bind="$attrs">
    <LogFilter
      v-model:params="state.params"
      :from-time="state.timeRange.fromTime"
      :to-time="state.timeRange.toTime"
      :loading="loading"
      :support-option-id-list="supportOptionIdList"
      @update:time="state.timeRange = $event"
    >
      <template #suffix>
        <NButton
          v-if="allowAdmin"
          type="default"
          @click="state.showSetting = true"
        >
          {{ $t("common.configure") }}
        </NButton>
        <NButton
          v-if="hasSyncPermission"
          type="primary"
          :disabled="!allowSync"
          :loading="syncing"
          @click="syncNow"
        >
          {{ $t("common.sync-now") }}
        </NButton>
      </template>
    </LogFilter>
    <div class="relative min-h-[8rem] w-full max-w-full overflow-auto">
      <LogTable
        :slow-query-log-list="slowQueryLogList"
        :show-placeholder="!loading"
        :show-project-column="showProjectColumn"
        :show-environment-column="showEnvironmentColumn"
        :show-instance-column="showInstanceColumn"
        :show-database-column="showDatabaseColumn"
        :allow-admin="allowAdmin"
        @select="selectSlowQueryLog"
      />
      <div
        v-if="loading"
        class="absolute inset-0 bg-white/50 flex flex-col items-center pt-[6rem]"
      >
        <BBSpin />
      </div>
    </div>

    <DetailPanel
      :slow-query-log="selectedSlowQueryLog"
      @close="selectedSlowQueryLog = undefined"
    />
  </div>

  <Drawer v-model:show="state.showSetting" width="auto">
    <DrawerContent
      :closable="true"
      class="w-[calc(100vw-2rem)] lg:max-w-[64rem] xl:max-w-[72rem]"
    >
      <SlowQuerySettings />
    </DrawerContent>
  </Drawer>
</template>

<script lang="ts" setup>
import dayjs from "dayjs";
import { NButton } from "naive-ui";
import { computed, shallowRef, watch, reactive } from "vue";
import { useI18n } from "vue-i18n";
import { BBSpin } from "@/bbkit";
import { Drawer, DrawerContent } from "@/components/v2";
import {
  pushNotification,
  useGracefulRequest,
  useSlowQueryPolicyList,
  useSlowQueryPolicyStore,
  useSlowQueryStore,
} from "@/store";
import type { ComposedSlowQueryLog, Permission } from "@/types";
import type { SearchScope, SearchParams, SearchScopeId } from "@/utils";
import { extractInstanceResourceName, hasWorkspacePermissionV2 } from "@/utils";
import { SlowQuerySettings } from "../Settings";
import DetailPanel from "./DetailPanel.vue";
import LogFilter from "./LogFilter.vue";
import LogTable from "./LogTable.vue";
import { buildListSlowQueriesRequest } from "./types";

const props = withDefaults(
  defineProps<{
    supportOptionIdList?: SearchScopeId[];
    readonlySearchScopes?: SearchScope[];
    showProjectColumn?: boolean;
    showEnvironmentColumn?: boolean;
    showInstanceColumn?: boolean;
    showDatabaseColumn?: boolean;
  }>(),
  {
    supportOptionIdList: () => ["environment", "project", "database"],
    readonlySearchScopes: () => [],
    showProjectColumn: true,
    showEnvironmentColumn: true,
    showInstanceColumn: true,
    showDatabaseColumn: true,
  }
);

const emit = defineEmits<{
  (event: "update:scopes", scopes: SearchScope[]): void;
}>();

interface LocalState {
  timeRange: {
    fromTime: number | undefined;
    toTime: number | undefined;
  };
  params: SearchParams;
  showSetting: boolean;
}

const defaultSlowQueryTimeRange = () => {
  const now = dayjs();
  const aWeekAgo = now.subtract(7, "days").startOf("day").valueOf();
  const tonight = now.endOf("day").valueOf();
  return {
    fromTime: aWeekAgo,
    toTime: tonight,
  };
};

const state = reactive<LocalState>({
  timeRange: defaultSlowQueryTimeRange(),
  params: {
    query: "",
    scopes: [],
  },
  showSetting: false,
});

const { t } = useI18n();
const slowQueryStore = useSlowQueryStore();
const loading = shallowRef(false);
const slowQueryLogList = shallowRef<ComposedSlowQueryLog[]>([]);
const selectedSlowQueryLog = shallowRef<ComposedSlowQueryLog>();
const syncing = shallowRef(false);

const hasSyncPermission = computed(() =>
  hasWorkspacePermissionV2("bb.instances.sync")
);

const searchScopes = computed(() => {
  return [...props.readonlySearchScopes, ...state.params.scopes];
});

watch(
  () => searchScopes.value,
  (scopes) => emit("update:scopes", scopes)
);

const selectedInstanceName = computed(() => {
  return searchScopes.value.find((s) => s.id === "instance")?.value;
});

const selectedProjectName = computed(() => {
  return searchScopes.value.find((s) => s.id === "project")?.value;
});

const params = computed(() => {
  const query = buildListSlowQueriesRequest(
    searchScopes.value,
    state.timeRange
  );
  return query;
});

const allowAdmin = computed(() => {
  const neededWorkspacePermissions: Permission[] = [
    "bb.instances.list",
    "bb.policies.update",
  ];
  return neededWorkspacePermissions.every((permission) =>
    hasWorkspacePermissionV2(permission)
  );
});

const { list: slowQueryPolicyList } = useSlowQueryPolicyList();

const allowSync = computed(() => {
  return (
    slowQueryPolicyList.value.filter((policy) => policy.slowQueryPolicy?.active)
      .length > 0
  );
});

const fetchSlowQueryLogList = async () => {
  loading.value = true;
  try {
    const list = await slowQueryStore.fetchSlowQueryLogList(params.value);
    slowQueryLogList.value = list;
  } finally {
    loading.value = false;
  }
};

const selectSlowQueryLog = (log: ComposedSlowQueryLog) => {
  selectedSlowQueryLog.value = log;
};

const syncNow = async () => {
  syncing.value = true;
  try {
    await useGracefulRequest(async () => {
      if (selectedInstanceName.value) {
        await slowQueryStore.syncSlowQueries(
          `instances/${selectedInstanceName.value}`
        );
      } else if (selectedProjectName.value) {
        await slowQueryStore.syncSlowQueries(
          `projects/${selectedProjectName.value}`
        );
      } else {
        const policyList = await useSlowQueryPolicyStore().fetchPolicyList();
        const requestList = policyList
          .filter((policy) => {
            return policy.slowQueryPolicy?.active;
          })
          .map(async (policy) => {
            return slowQueryStore.syncSlowQueries(
              `instances/${extractInstanceResourceName(policy.name)}`
            );
          });
        await Promise.all(requestList);
      }

      pushNotification({
        module: "devsecdb",
        style: "SUCCESS",
        title: t("slow-query.sync-job-started"),
      });
    });
  } finally {
    syncing.value = false;
  }
};

// Fetch the list while params changed.
watch(() => JSON.stringify(params.value), fetchSlowQueryLogList, {
  immediate: true,
});
</script>
