//#region src/features/specification-extension/helpers.ts
/**
* Utility function to extract all keys starting with 'x-' (OpenAPI extensions) from an object.
*
* @param object - The object from which to extract extension keys.
* @returns An object containing only the entries whose keys start with 'x-'.
*/
var getXKeysFromObject = (object) => {
	if (!object) return {};
	return Object.fromEntries(Object.entries(object).filter(([key]) => key.startsWith("x-")));
};
//#endregion
export { getXKeysFromObject };

//# sourceMappingURL=helpers.js.map