<template>
  <div class="mx-auto w-full max-w-sm">
    <DevsecdbLogo class="mx-auto" />

    <div class="mt-8">
      <NCard>
        <p class="text-xl pl-1 font-medium mb-4">
          {{ $t("common.sign-in-as-admin") }}
        </p>
        <PasswordSigninForm :show-forgot-password="false" />
      </NCard>
    </div>
  </div>
  <AuthFooter />
</template>

<script lang="ts" setup>
import { NCard } from "naive-ui";
import DevsecdbLogo from "@/components/DevsecdbLogo.vue";
import PasswordSigninForm from "@/components/PasswordSigninForm.vue";
import AuthFooter from "./AuthFooter.vue";
</script>
