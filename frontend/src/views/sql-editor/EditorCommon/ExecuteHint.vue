<template>
  <div class="execute-hint w-112">
    <NAlert type="info">
      <section class="space-y-2">
        <p>
          <i18n-t keypath="sql-editor.only-select-allowed">
            <template #select>
              <strong
                ><code>SELECT</code>, <code>SHOW</code> and
                <code>SET</code></strong
              >
            </template>
          </i18n-t>
        </p>
        <p v-if="database">
          <i18n-t keypath="sql-editor.enable-ddl-for-environment">
            <template #environment>
              <EnvironmentV1Name
                :environment="database.effectiveEnvironmentEntity"
              />
            </template>
          </i18n-t>
        </p>
        <p v-if="descriptions.action && descriptions.reaction">
          <i18n-t keypath="sql-editor.want-to-action">
            <template #want>
              {{ descriptions.want }}
            </template>
            <template #action>
              <strong>
                {{ descriptions.action }}
              </strong>
            </template>
            <template #reaction>
              {{ descriptions.reaction }}
            </template>
          </i18n-t>
        </p>
      </section>
    </NAlert>

    <div class="execute-hint-content mt-4 flex justify-between">
      <div
        v-if="actions.issue && actions.admin"
        class="flex justify-start items-center space-x-2"
      >
        <AdminModeButton @enter="$emit('close')" />
      </div>
      <div class="flex flex-1 justify-end items-center space-x-2">
        <NButton @click="handleClose">{{ $t("common.close") }}</NButton>
        <NButton
          v-if="actions.issue"
          type="primary"
          @click="handleClickCreateIssue"
        >
          {{ descriptions.action }}
        </NButton>
        <AdminModeButton v-else @enter="$emit('close')" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computedAsync } from "@vueuse/core";
import { NAlert, NButton } from "naive-ui";
import { v4 as uuidv4 } from "uuid";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { parseSQL, isDDLStatement } from "@/components/MonacoEditor/sqlParser";
import { EnvironmentV1Name } from "@/components/v2";
import { PROJECT_V1_ROUTE_ISSUE_DETAIL } from "@/router/dashboard/projectV1";
import {
  pushNotification,
  useAppFeature,
  useDatabaseV1Store,
  useSQLEditorTabStore,
} from "@/store";
import type { ComposedDatabase } from "@/types";
import { extractProjectResourceName, hasWorkspacePermissionV2 } from "@/utils";
import AdminModeButton from "./AdminModeButton.vue";

withDefaults(
  defineProps<{
    database?: ComposedDatabase | undefined;
  }>(),
  { database: undefined }
);

const emit = defineEmits<{
  (e: "close"): void;
}>();

const DDLIssueTemplate = "bb.issue.database.schema.update";
const DMLIssueTemplate = "bb.issue.database.data.update";

const router = useRouter();
const { t } = useI18n();
const tabStore = useSQLEditorTabStore();
const disallowNavigateToConsole = useAppFeature(
  "bb.feature.disallow-navigate-to-console"
);

const statement = computed(() => {
  const tab = tabStore.currentTab;
  return tab?.selectedStatement || tab?.statement || "";
});

const isDDL = computedAsync(async () => {
  const { data } = await parseSQL(statement.value);
  return data !== null ? isDDLStatement(data, "some") : false;
}, false);

const actions = computed(() => {
  type Actions = {
    admin: boolean;
    issue: boolean;
  };
  const actions: Actions = {
    admin: false,
    issue: false,
  };
  if (hasWorkspacePermissionV2("bb.sql.admin")) {
    actions.admin = true;
  }
  if (!disallowNavigateToConsole.value) {
    actions.issue = true;
  }

  return actions;
});

const descriptions = computed(() => {
  const descriptions = {
    want: isDDL.value
      ? t("database.edit-schema").toLowerCase()
      : t("database.change-data").toLowerCase(),
    action: "",
    reaction: "",
  };
  const { admin, issue } = actions.value;
  if (issue) {
    descriptions.action = isDDL.value
      ? t("database.edit-schema")
      : t("database.change-data");
    descriptions.reaction = t("sql-editor.and-submit-an-issue");
  } else if (admin) {
    descriptions.action = t("sql-editor.admin-mode.self");
    descriptions.reaction = t("sql-editor.to-enable-admin-mode");
  }
  return descriptions;
});

const handleClose = () => {
  emit("close");
};

const gotoCreateIssue = () => {
  const database = tabStore.currentTab?.connection.database ?? "";
  if (!database) {
    pushNotification({
      module: "devsecdb",
      style: "CRITICAL",
      title: t("sql-editor.goto-edit-schema-hint"),
    });
    return;
  }

  emit("close");

  const db = useDatabaseV1Store().getDatabaseByName(database);
  const sqlStorageKey = `bb.issues.sql.${uuidv4()}`;
  localStorage.setItem(sqlStorageKey, statement.value);
  const route = router.resolve({
    name: PROJECT_V1_ROUTE_ISSUE_DETAIL,
    params: {
      projectId: extractProjectResourceName(db.project),
      issueSlug: "create",
    },
    query: {
      template: isDDL.value ? DDLIssueTemplate : DMLIssueTemplate,
      name: `[${db.databaseName}] ${
        isDDL.value ? "Edit schema" : "Change Data"
      }`,
      databaseList: db.name,
      sqlStorageKey,
    },
  });
  window.open(route.fullPath, "_blank");
};

const handleClickCreateIssue = () => {
  gotoCreateIssue();
};
</script>
