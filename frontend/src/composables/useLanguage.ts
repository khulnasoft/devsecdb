import { useLocalStorage } from "@vueuse/core";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { type LanguageStorage } from "@/plugins/i18n";
import { emitStorageChangedEvent } from "@/utils";

/**
 * Language hook for i18n
 * @returns
 */
const useLanguage = () => {
  const { availableLocales, locale } = useI18n();
  const currentRoute = useRouter().currentRoute;
  const storage = useLocalStorage<LanguageStorage>("devsecdb_options", {
    appearance: {
      language: "",
    },
  });

  const setLocale = (lang: string) => {
    locale.value = lang;
    storage.value = {
      appearance: {
        language: lang,
      },
    };
    emitStorageChangedEvent();

    if (currentRoute.value.meta.title) {
      const title = currentRoute.value.meta.title(currentRoute.value);
      document.title = title || "Devsecdb";
    }
  };

  const toggleLocales = () => {
    const locales = availableLocales;
    const nextLocale =
      locales[(locales.indexOf(locale.value) + 1) % locales.length];
    setLocale(nextLocale);
  };

  return {
    locale,
    availableLocales,
    setLocale,
    toggleLocales,
  };
};

export { useLanguage };
