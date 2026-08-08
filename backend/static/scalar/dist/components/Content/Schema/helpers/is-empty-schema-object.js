import { isTypeObject } from "./is-type-object.js";
//#region src/components/Content/Schema/helpers/is-empty-schema-object.ts
/**
* Determines if the given schema is an empty object schema.
* An empty object schema is defined as a schema with type 'object'
* and no defined properties, no additionalProperties (or set to false), and no patternProperties.
*/
var isEmptySchemaObject = (schema) => {
	if (!isTypeObject(schema)) return false;
	const hasNoProperties = Object.keys(schema.properties ?? {}).length === 0;
	const hasNoAdditionalProperties = schema.additionalProperties === void 0 || schema.additionalProperties === false;
	const hasNoPatternProperties = Object.keys(schema.patternProperties ?? {}).length === 0;
	return hasNoProperties && hasNoAdditionalProperties && hasNoPatternProperties;
};
//#endregion
export { isEmptySchemaObject };

//# sourceMappingURL=is-empty-schema-object.js.map