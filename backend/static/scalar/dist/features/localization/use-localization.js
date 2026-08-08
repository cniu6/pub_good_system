import { RTL_LOCALES, localeTranslations } from "./translations.js";
import { computed, inject, provide, toValue } from "vue";
import { isObject } from "@scalar/helpers/object/is-object";
import { mergeObjects } from "@scalar/helpers/object/merge-objects";
//#region src/features/localization/use-localization.ts
var LOCALIZATION_SYMBOL = Symbol("LOCALIZATION");
var resolveBuiltInLocale = (locale) => {
	if (!locale) return "en";
	if (locale in localeTranslations) return locale;
	const normalized = locale.replace("_", "-").toLowerCase();
	if (normalized.startsWith("zh")) return "zh-CN";
	const language = normalized.split("-")[0];
	return Object.keys(localeTranslations).find((key) => key.toLowerCase() === language) ?? "en";
};
var resolveDirection = (localization) => {
	if (localization?.direction && localization.direction !== "auto") return localization.direction;
	const language = (localization?.locale ?? "en").replace("_", "-").split("-")[0]?.toLowerCase();
	return language && RTL_LOCALES.has(language) ? "rtl" : "ltr";
};
var resolveLocalization = (localization) => {
	const locale = localization?.locale ?? "en";
	const builtInLocale = resolveBuiltInLocale(locale);
	const translations = mergeObjects(mergeObjects(localeTranslations.en, localeTranslations[builtInLocale]), localization?.translations);
	return {
		locale,
		direction: resolveDirection(localization),
		translations
	};
};
var getTranslationValue = (translations, key) => key.split(".").reduce((value, segment) => {
	if (!isObject(value)) return;
	return value[segment];
}, translations);
var translateApiReference = (translations, key, params) => {
	const value = getTranslationValue(translations, key);
	const template = typeof value === "string" ? value : key;
	if (!params) return template;
	return Object.entries(params).reduce((result, [param, paramValue]) => result.replaceAll(`{${param}}`, String(paramValue)), template);
};
var createLocalizationContext = (localization) => {
	const resolved = computed(() => resolveLocalization(toValue(localization)));
	return {
		locale: computed(() => resolved.value.locale),
		direction: computed(() => resolved.value.direction),
		translations: computed(() => resolved.value.translations),
		translate: (key, params) => translateApiReference(resolved.value.translations, key, params)
	};
};
var provideLocalization = (localization) => {
	const context = createLocalizationContext(localization);
	provide(LOCALIZATION_SYMBOL, context);
	return context;
};
var useLocalization = () => inject(LOCALIZATION_SYMBOL, createLocalizationContext(void 0));
//#endregion
export { provideLocalization, resolveLocalization, useLocalization };

//# sourceMappingURL=use-localization.js.map