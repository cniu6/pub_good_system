import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { isObjectSchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/features/Operation/helpers/flatten-deep-object-query-parameter.ts
var isParameterWithSchema = (parameter) => "schema" in parameter && parameter.schema !== void 0;
var resolveSchema = (schema) => {
	return getResolvedRef(schema);
};
var toFlattenedDeepObjectParameter = (parameter, name, description, required, schema) => {
	const { example: _example, examples: _examples, ...parameterWithoutExamples } = parameter;
	return {
		...parameterWithoutExamples,
		name,
		description,
		required,
		schema
	};
};
var flattenDeepObjectProperties = (parameter, schema, namePrefix) => {
	if (!schema.properties) return [parameter];
	const requiredProperties = new Set(schema.required ?? []);
	const flattenedParameters = Object.entries(schema.properties).flatMap(([propertyName, propertySchema]) => {
		const resolvedPropertySchema = resolveSchema(propertySchema);
		if (!resolvedPropertySchema) return [];
		const nestedName = `${namePrefix}[${propertyName}]`;
		const nestedParameter = toFlattenedDeepObjectParameter(parameter, nestedName, resolvedPropertySchema.description ?? parameter.description, requiredProperties.has(propertyName), resolvedPropertySchema);
		if (isObjectSchema(resolvedPropertySchema) && resolvedPropertySchema.properties) return flattenDeepObjectProperties(nestedParameter, resolvedPropertySchema, nestedName);
		return [nestedParameter];
	});
	return flattenedParameters.length > 0 ? flattenedParameters : [parameter];
};
/**
* Deep object query parameters serialize as name[prop] pairs in URLs.
* Rendering the same shape keeps docs aligned with the request UI.
*/
var flattenDeepObjectQueryParameter = (parameter) => {
	if (parameter.in !== "query" || !isParameterWithSchema(parameter) || parameter.style !== "deepObject") return [parameter];
	const resolvedSchema = resolveSchema(parameter.schema);
	if (!resolvedSchema || !isObjectSchema(resolvedSchema)) return [parameter];
	return flattenDeepObjectProperties(parameter, resolvedSchema, parameter.name);
};
//#endregion
export { flattenDeepObjectQueryParameter };

//# sourceMappingURL=flatten-deep-object-query-parameter.js.map