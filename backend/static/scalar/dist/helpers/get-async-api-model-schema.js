import { unwrapAsyncApiSchema } from "./get-async-api-message-payload-schema.js";
import { getResolvedRef, mergeSiblingReferences } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/helpers/get-async-api-model-schema.ts
/**
* Resolves a named schema from an AsyncAPI document's `components.schemas` into the `SchemaObject`
* shape the shared OpenAPI Model components and search indexing expect.
*
* AsyncAPI keeps reusable schemas in the same place OpenAPI does, but an entry may need extra handling:
* - `components` or the schema itself may be a `$ref`; siblings are merged so keys declared alongside
*   a `$ref` are kept rather than dropped.
* - A Multi Format Schema Object (`schemaFormat` plus a nested `schema`) is unwrapped to its payload,
*   so we index and render the inner JSON Schema rather than the wrapper.
* - Boolean (`true`/`false`) and non-JSON-Schema payloads are skipped.
*/
var getAsyncApiModelSchema = (document, name) => {
	if (!document.components) return;
	const entry = getResolvedRef(document.components, mergeSiblingReferences).schemas?.[name];
	return unwrapAsyncApiSchema(entry);
};
//#endregion
export { getAsyncApiModelSchema };

//# sourceMappingURL=get-async-api-model-schema.js.map