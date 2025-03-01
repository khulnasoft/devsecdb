<template>
  <div class="flex flex-col gap-y-4">
    <VCSProviderBasicInfoPanel :create="false" :config="state.config" />

    <div class="pt-4 mt-2 flex border-t justify-between">
      <template v-if="connectorList.length == 0">
        <BBButtonConfirm
          :type="'DELETE'"
          :button-text="$t('gitops.setting.git-provider.delete')"
          :ok-text="$t('common.delete')"
          :confirm-title="
            $t('gitops.setting.git-provider.delete-confirm', {
              name: vcs?.title,
            })
          "
          :disabled="!hasDeleteVCSPermission"
          :require-confirm="true"
          @confirm="deleteVCS"
        />
      </template>
      <template v-else>
        <div class="mt-1 textinfolabel">
          {{ $t("gitops.setting.git-provider.delete-forbidden") }}
        </div>
      </template>
      <div class="space-x-3">
        <NButton
          v-if="allowUpdate"
          :disabled="state.loading"
          @click.prevent="cancel"
        >
          {{ $t("common.discard-changes") }}
        </NButton>
        <NButton
          type="primary"
          :disabled="!allowUpdate"
          :loading="state.loading"
          @click.prevent="doUpdate"
        >
          {{ $t("common.update") }}
        </NButton>
      </div>
    </div>
  </div>

  <div class="py-6 space-y-4">
    <div class="text-lg leading-6 font-medium text-main">
      {{ $t("repository.linked") + ` (${connectorList.length})` }}
    </div>
    <VCSConnectorTable
      v-if="hasListRepoPermission"
      :connector-list="connectorList"
    />
    <NoPermissionPlaceholder v-else />
  </div>
</template>

<script lang="ts" setup>
import isEmpty from "lodash-es/isEmpty";
import { NButton } from "naive-ui";
import { reactive, computed, watchEffect } from "vue";
import { useRouter } from "vue-router";
import { BBButtonConfirm } from "@/bbkit";
import { VCSProviderBasicInfoPanel } from "@/components/VCS";
import VCSConnectorTable from "@/components/VCSConnectorTable.vue";
import NoPermissionPlaceholder from "@/components/misc/NoPermissionPlaceholder.vue";
import { WORKSPACE_ROUTE_GITOPS } from "@/router/dashboard/workspaceRoutes";
import {
  pushNotification,
  useVCSConnectorStore,
  useVCSProviderStore,
} from "@/store";
import type { VCSConfig } from "@/types";
import { VCSType } from "@/types/proto/v1/common";
import type { VCSProvider } from "@/types/proto/v1/vcs_provider_service";
import { hasWorkspacePermissionV2 } from "@/utils";

interface LocalState {
  config: VCSConfig;
  loading: boolean;
}

const props = defineProps<{
  vcsResourceId: string;
}>();

const router = useRouter();
const vcsV1Store = useVCSProviderStore();
const vcsConnectorStore = useVCSConnectorStore();

const vcs = computed((): VCSProvider | undefined => {
  return vcsV1Store.getVCSById(props.vcsResourceId);
});

const initState = computed(
  (): VCSConfig => ({
    type: vcs.value?.type ?? VCSType.GITLAB,
    uiType: "GITLAB_SELF_HOST",
    resourceId: props.vcsResourceId,
    name: vcs.value?.title ?? "",
    instanceUrl: vcs.value?.url ?? "",
    accessToken: "",
  })
);

const resetState = () => {
  state.config = { ...initState.value };
  state.loading = false;
};

const state = reactive<LocalState>({
  config: { ...initState.value },
  loading: false,
});

const hasUpdateVCSPermission = computed(() => {
  return hasWorkspacePermissionV2("bb.vcsProviders.update");
});

const hasDeleteVCSPermission = computed(() => {
  return hasWorkspacePermissionV2("bb.vcsProviders.delete");
});

const hasListRepoPermission = computed(() => {
  return hasWorkspacePermissionV2("bb.vcsProviders.listProjects");
});

watchEffect(async () => {
  await vcsV1Store.getOrFetchVCSList();
  resetState();
  if (vcs.value && hasListRepoPermission.value) {
    await vcsConnectorStore.fetchConnectorsInProvider(vcs.value.name);
  }
});

const connectorList = computed(() => {
  return vcsConnectorStore.getConnectorsInProvider(vcs.value?.name ?? "");
});

const allowUpdate = computed(() => {
  return (
    (state.config.name != vcs.value?.title ||
      !isEmpty(state.config.accessToken)) &&
    hasUpdateVCSPermission.value
  );
});

const doUpdate = async () => {
  if (!vcs.value) {
    return;
  }
  state.loading = true;

  const vcsPatch: Partial<VCSProvider> = {
    name: vcs.value.name,
  };
  if (state.config.name != vcs.value.title) {
    vcsPatch.title = state.config.name;
  }
  if (!isEmpty(state.config.accessToken)) {
    vcsPatch.accessToken = state.config.accessToken;
  }

  try {
    const updatedVCS = await vcsV1Store.updateVCS(vcsPatch);
    if (!updatedVCS) {
      return;
    }
    resetState();
    pushNotification({
      module: "devsecdb",
      style: "SUCCESS",
      title: `Successfully updated '${updatedVCS.title}'`,
    });
  } finally {
    state.loading = false;
  }
};

const cancel = () => {
  resetState();
};

const deleteVCS = async () => {
  if (!vcs.value) {
    return;
  }
  state.loading = true;

  try {
    const title = vcs.value.title;
    await vcsV1Store.deleteVCS(vcs.value.name);
    pushNotification({
      module: "devsecdb",
      style: "SUCCESS",
      title: `Successfully deleted '${title}'`,
    });
    router.push({
      name: WORKSPACE_ROUTE_GITOPS,
    });
  } finally {
    state.loading = false;
  }
};
</script>
