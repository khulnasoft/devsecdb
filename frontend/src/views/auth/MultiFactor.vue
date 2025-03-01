<template>
  <div
    class="mx-auto w-full h-full py-6 flex flex-col justify-center items-center bg-gray-100 rounded-lg"
  >
    <div class="w-80 bg-white p-8 py-6 rounded-lg shadow">
      <img
        class="h-12 w-auto mx-auto mb-8"
        src="@/assets/logo-full.svg"
        alt="Devsecdb"
      />
      <form
        class="w-full mt-4 h-auto flex flex-col justify-start items-center"
        @submit.prevent="challenge"
      >
        <template v-if="state.selectedMFAType === 'OTP'">
          <heroicons-outline:device-phone-mobile
            class="w-8 h-auto opacity-60"
          />
          <p class="my-2 mb-4">{{ $t("multi-factor.auth-code") }}</p>
          <BBTextField
            v-model:value="state.otpCode"
            placeholder="XXXXXX"
            class="w-full"
          />
        </template>
        <template v-else-if="state.selectedMFAType === 'RECOVERY_CODE'">
          <heroicons-outline:key class="w-8 h-auto opacity-60" />
          <p class="my-2 mb-4">{{ $t("multi-factor.recovery-code") }}</p>
          <BBTextField
            v-model:value="state.recoveryCode"
            placeholder="XXXXXXXXXX"
            class="w-full"
          />
        </template>
        <div class="w-full mt-4">
          <NButton class="!w-full" attr-type="submit" type="primary">
            {{ $t("common.verify") }}
          </NButton>
        </div>
        <p class="textinfolabel mt-2">
          {{ challengeDescription }}
        </p>
      </form>
      <hr class="my-3" />
      <div class="text-sm mb-2">
        <p class="">{{ $t("multi-factor.other-methods.self") }}:</p>
        <ul class="list-disc list-inside pl-2 pt-1">
          <li v-if="state.selectedMFAType !== 'OTP'">
            <button class="accent-link" @click="state.selectedMFAType = 'OTP'">
              {{ $t("multi-factor.other-methods.use-auth-app.self") }}
            </button>
          </li>
          <li v-if="state.selectedMFAType !== 'RECOVERY_CODE'">
            <button
              class="accent-link"
              @click="state.selectedMFAType = 'RECOVERY_CODE'"
            >
              {{ $t("multi-factor.other-methods.use-recovery-code.self") }}
            </button>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { NButton } from "naive-ui";
import { computed, reactive } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { BBTextField } from "@/bbkit";
import { useAuthStore } from "@/store";

type MFAType = "OTP" | "RECOVERY_CODE";

interface LocalState {
  selectedMFAType: MFAType;
  otpCode: string;
  recoveryCode: string;
}

const { t } = useI18n();
const route = useRoute();
const authStore = useAuthStore();
const state = reactive<LocalState>({
  selectedMFAType: "OTP",
  otpCode: "",
  recoveryCode: "",
});

const mfaTempToken = computed(() => {
  return route.query.mfaTempToken as string;
});

const challengeDescription = computed(() => {
  if (state.selectedMFAType === "OTP") {
    return t("multi-factor.other-methods.use-auth-app.description");
  } else if (state.selectedMFAType === "RECOVERY_CODE") {
    return t("multi-factor.other-methods.use-recovery-code.description");
  } else {
    return "";
  }
});

const challenge = async () => {
  const mfaContext: any = {};
  if (state.selectedMFAType === "OTP") {
    mfaContext.otpCode = state.otpCode;
  } else if (state.selectedMFAType === "RECOVERY_CODE") {
    mfaContext.recoveryCode = state.recoveryCode;
  }
  await authStore.login({
    web: true,
    mfaTempToken: mfaTempToken.value,
    ...mfaContext,
  });
};
</script>
