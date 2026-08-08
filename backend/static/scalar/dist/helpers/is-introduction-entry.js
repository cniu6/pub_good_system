//#region src/helpers/is-introduction-entry.ts
/**
* The auto-generated "Introduction" navigation entry always uses the `introduction` slug, so its id
* ends with this suffix regardless of the active locale or the displayed title.
*/
var INTRODUCTION_ENTRY_ID_SUFFIX = "/description/introduction";
/**
* Whether a navigation entry is the auto-generated introduction section.
*
* We match on the stable entry id instead of the English title so localization does not accidentally
* rewrite unrelated description headings that happen to be called "Introduction".
*/
var isIntroductionEntry = (entry) => entry.type === "text" && typeof entry.id === "string" && entry.id.endsWith("/description/introduction");
//#endregion
export { INTRODUCTION_ENTRY_ID_SUFFIX, isIntroductionEntry };

//# sourceMappingURL=is-introduction-entry.js.map