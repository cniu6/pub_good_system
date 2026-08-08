import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { isObject } from "@scalar/helpers/object/is-object";
//#region src/helpers/get-async-api-message-payload-schema.ts
/**
* A resolved schema-bearing value may be a JSON Schema object, a boolean (`true`/`false`)
* schema, or a Multi Format Schema wrapper, so accept only plain objects and treat them as the
* `SchemaObject` the shared Schema/Model rendering expects.
*/
var isSchemaObject = (value) => isObject(value);
/**
* Unwraps an AsyncAPI schema-bearing value into the `SchemaObject` shape the shared OpenAPI
* `Schema`/`Model` components expect.
*
* AsyncAPI reuses JSON Schema, but a value may need extra handling:
* - The value itself may be a `$ref`; it is resolved to its `$ref-value`.
* - A Multi Format Schema Object (`schemaFormat` plus a nested `schema`) is unwrapped to its
*   payload, so we render the inner JSON Schema rather than the wrapper.
* - Boolean (`true`/`false`) and non-JSON-Schema payloads are skipped.
*/
var unwrapAsyncApiSchema = (value) => {
	if (value === void 0) return;
	const resolved = getResolvedRef(value);
	const schema = isObject(resolved) && "schemaFormat" in resolved ? getResolvedRef(resolved.schema) : resolved;
	return isSchemaObject(schema) ? schema : void 0;
};
/** Resolves an AsyncAPI message `payload` into the `SchemaObject` the shared Schema component expects. */
var getAsyncApiMessagePayloadSchema = (message) => unwrapAsyncApiSchema(message.payload);
/** Resolves an AsyncAPI message `headers` into the `SchemaObject` the shared Schema component expects. */
var getAsyncApiMessageHeadersSchema = (message) => unwrapAsyncApiSchema(message.headers);
//#endregion
export { getAsyncApiMessageHeadersSchema, getAsyncApiMessagePayloadSchema, unwrapAsyncApiSchema };

//# sourceMappingURL=get-async-api-message-payload-schema.js.map