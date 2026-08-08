//#region src/components/Content/AsyncApi/helpers/filter-children-by-type.ts
/**
* Narrow a navigation entry's children down to a single AsyncAPI node type.
*
* The channel renderer needs its operations and the operation renderer needs its messages; both
* share this type-guarded filter so the cast lives in one place.
*/
var filterChildrenByType = (children, type) => (children ?? []).filter((child) => child.type === type);
//#endregion
export { filterChildrenByType };

//# sourceMappingURL=filter-children-by-type.js.map