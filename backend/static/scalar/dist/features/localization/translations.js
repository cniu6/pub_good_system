import { ar } from "./locales/ar.js";
import { de } from "./locales/de.js";
import { en } from "./locales/en.js";
import { es } from "./locales/es.js";
import { fr } from "./locales/fr.js";
import { pt } from "./locales/pt.js";
import { ru } from "./locales/ru.js";
import { zhCn } from "./locales/zh-cn.js";
//#region src/features/localization/translations.ts
var localeTranslations = {
	en,
	ru,
	es,
	fr,
	de,
	"zh-CN": zhCn,
	ar,
	pt
};
var RTL_LOCALES = /* @__PURE__ */ new Set([
	"ar",
	"fa",
	"he",
	"ur"
]);
//#endregion
export { RTL_LOCALES, localeTranslations };

//# sourceMappingURL=translations.js.map