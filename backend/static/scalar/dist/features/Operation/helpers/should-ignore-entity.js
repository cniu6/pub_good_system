//#region src/features/Operation/helpers/should-ignore-entity.ts
/**
* Check if an entity should be ignored
*/
var shouldIgnoreEntity = (data) => Boolean(data?.["x-internal"] || data?.["x-scalar-ignore"]);
//#endregion
export { shouldIgnoreEntity };

//# sourceMappingURL=should-ignore-entity.js.map