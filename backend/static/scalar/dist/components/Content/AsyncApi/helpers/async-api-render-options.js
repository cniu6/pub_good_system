//#region src/components/Content/AsyncApi/helpers/async-api-render-options.ts
/**
* Fill in defaults so the shared renderers always receive a complete ordering options object,
* regardless of which fields the caller provided.
*/
var resolveSchemaRenderOptions = (options) => ({
	orderRequiredPropertiesFirst: options?.orderRequiredPropertiesFirst ?? false,
	orderSchemaPropertiesBy: options?.orderSchemaPropertiesBy ?? "preserve",
	expandAllSchemaProperties: options?.expandAllSchemaProperties ?? false
});
//#endregion
export { resolveSchemaRenderOptions };

//# sourceMappingURL=async-api-render-options.js.map