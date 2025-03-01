<template>
  <div>
    <p class="font-medium flex flex-row justify-start items-center">
      <span class="mr-2">{{
        $t("settings.general.workspace.maximum-role-expiration.self")
      }}</span>
    </p>
    <p class="text-sm text-gray-400 mt-1">
      {{ $t("settings.general.workspace.maximum-role-expiration.description") }}
    </p>
    <div class="mt-3 w-full flex flex-row justify-start items-center gap-4">
      <NInputNumber
        v-model:value="state.inputValue"
        class="w-40"
        :disabled="!allowChangeSetting || state.neverExpire"
        :min="1"
        :precision="0"
      >
        <template #suffix>
          {{ $t("settings.general.workspace.maximum-role-expiration.days") }}
        </template>
      </NInputNumber>
      <NCheckbox v-model:checked="state.neverExpire" style="margin-right: 12px">
        {{
          $t("settings.general.workspace.maximum-role-expiration.never-expires")
        }}
      </NCheckbox>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useDebounceFn } from "@vueuse/core";
import { NInputNumber, NCheckbox } from "naive-ui";
import { computed, reactive, watch } from "vue";
import { useI18n } from "vue-i18n";
import { pushNotification } from "@/store";
import { useSettingV1Store } from "@/store/modules/v1/setting";
import { Duration } from "@/types/proto/google/protobuf/duration";

const DEFAULT_EXPIRATION_DAYS = 90;

const getInitialState = (): LocalState => {
  const defaultState: LocalState = {
    inputValue: DEFAULT_EXPIRATION_DAYS,
    neverExpire: true,
  };
  const seconds =
    settingV1Store.workspaceProfileSetting?.maximumRoleExpiration?.seconds?.toNumber();
  if (seconds && seconds > 0) {
    defaultState.inputValue =
      Math.floor(seconds / (60 * 60 * 24)) || DEFAULT_EXPIRATION_DAYS;
    defaultState.neverExpire = false;
  }
  return defaultState;
};

interface LocalState {
  inputValue: number;
  neverExpire: boolean;
}

const props = defineProps<{
  allowEdit: boolean;
}>();

const { t } = useI18n();
const settingV1Store = useSettingV1Store();
const state = reactive<LocalState>(getInitialState());

const allowChangeSetting = computed(() => {
  return props.allowEdit;
});

const handleSettingChange = useDebounceFn(async () => {
  let seconds = -1;
  if (!state.neverExpire) {
    seconds = state.inputValue * 24 * 60 * 60;
  }
  await settingV1Store.updateWorkspaceProfile({
    payload: {
      maximumRoleExpiration: Duration.fromPartial({ seconds, nanos: 0 }),
    },
    updateMask: [
      "value.workspace_profile_setting_value.maximum_role_expiration",
    ],
  });
  pushNotification({
    module: "devsecdb",
    style: "SUCCESS",
    title: t("settings.general.workspace.config-updated"),
  });
}, 2000);

watch(
  () => [state.inputValue, state.neverExpire],
  () => {
    handleSettingChange();
  }
);
</script>
