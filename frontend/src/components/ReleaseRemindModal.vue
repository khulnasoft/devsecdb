<template>
  <BBModal
    :title="$t('settings.release.new-version-available')"
    @close="$emit('cancel')"
  >
    <div class="min-w-0 md:min-w-400">
      <div>
        <p class="whitespace-pre-wrap">
          <i18n-t keypath="settings.release.new-version-content">
            <template #tag>
              <a
                class="font-bold underline"
                target="_blank"
                :href="actuatorStore.releaseInfo.latest?.html_url"
              >
                {{ actuatorStore.releaseInfo.latest?.tag_name }}
              </a>
            </template>
          </i18n-t>
        </p>
        <NCheckbox
          class="mt-3 ml-1"
          :label="$t('settings.release.not-show-till-next-release')"
          :checked="actuatorStore.releaseInfo.ignoreRemindModalTillNextRelease"
          @update:checked="
            (on: boolean) =>
              (actuatorStore.releaseInfo.ignoreRemindModalTillNextRelease = on)
          "
        />
      </div>
      <div class="mt-7 flex justify-end space-x-2">
        <NButton @click="$emit('cancel')">
          {{ $t("common.dismiss") }}
        </NButton>
        <NButton type="primary" @click="onClick">
          {{ $t("common.learn-more") }}
        </NButton>
      </div>
    </div>
  </BBModal>
</template>

<script lang="ts" setup>
import { NButton, NCheckbox } from "naive-ui";
import { computed } from "vue";
import { BBModal } from "@/bbkit";
import { useActuatorV1Store, useSubscriptionV1Store } from "@/store";

const emit = defineEmits(["cancel"]);

const actuatorStore = useActuatorV1Store();
const subscriptionStore = useSubscriptionV1Store();

const link = computed(() => {
  if (subscriptionStore.isSelfHostLicense) {
    return "https://www.secdb.khulnasoft.com/docs/get-started/install/overview";
  }
  return subscriptionStore.purchaseLicenseUrl;
});

const onClick = () => {
  window.open(link.value, "_blank");
  emit("cancel");
};
</script>
