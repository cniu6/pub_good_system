import { REGEX } from "@scalar/helpers/regex/regex-helpers";
//#region src/components/Content/Schema/helpers/get-ref-name.ts
/**
* Gets the "name" of the schema from the ref path
* TODO: this will change so fix it when the new refs are out
* Then add tests
*
* @example SchemaName from #/components/schemas/SchemaName
*/
var getRefName = (ref) => {
	if (!ref) return null;
	const match = ref.match(REGEX.REF_NAME);
	if (match) return match[1];
	return null;
};
//#endregion
export { getRefName };

//# sourceMappingURL=get-ref-name.js.map