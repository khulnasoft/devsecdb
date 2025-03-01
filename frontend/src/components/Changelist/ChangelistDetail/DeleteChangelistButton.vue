<template>
  <NTooltip :disabled="errors.length === 0">
    <template #trigger>
      <NPopconfirm :disabled="errors.length > 0" @positive-click="handleDelete">
        <template #trigger>
          <NButton
            quaternary
            size="small"
            style="--n-padding: 0 6px"
            :disabled="errors.length > 0"
            @click.stop
          >
            <template #icon>
              <heroicons:trash />
            </template>
            <template #default>
              {{ $t("changelist.delete-this-changelist") }}
            </template>
          </NButton>
        </template>

        <template #default>
          <div>{{ $t("changelist.confirm-delete-changelist") }}</div>
        </template>
      </NPopconfirm>
    </template>

    <template #default>
      <ErrorList :errors="errors" />
    </template>
  </NTooltip>
</template>

<script setup lang="ts">
import { NButton, NPopconfirm, NTooltip } from "naive-ui";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import ErrorList from "@/components/misc/ErrorList.vue";
import { PROJECT_V1_ROUTE_CHANGELISTS } from "@/router/dashboard/projectV1";
import { pushNotification, useChangelistStore } from "@/store";
import { useChangelistDetailContext } from "./context";

const router = useRouter();
const { t } = useI18n();
const { allowEdit, changelist, isUpdating } = useChangelistDetailContext();

const errors = computed(() => {
  const errors: string[] = [];
  if (!allowEdit.value) {
    errors.push(
      t("changelist.error.you-are-not-allowed-to-perform-this-action")
    );
    return errors;
  }
  return errors;
});

const handleDelete = async () => {
  isUpdating.value = true;
  try {
    await useChangelistStore().deleteChangelist(changelist.value.name);
    router.replace({
      name: PROJECT_V1_ROUTE_CHANGELISTS,
    });
    pushNotification({
      module: "devsecdb",
      style: "SUCCESS",
      title: t("common.deleted"),
    });
  } catch {
    isUpdating.value = false;
  }
};
</script>
