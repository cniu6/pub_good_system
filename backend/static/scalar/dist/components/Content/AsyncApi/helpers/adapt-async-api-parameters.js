import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/components/Content/AsyncApi/helpers/adapt-async-api-parameters.ts
/**
* Adapt AsyncAPI channel parameters into the OpenAPI `ParameterObject` shape so we can reuse the
* existing `ParameterList` rendering instead of building a separate component.
*
* AsyncAPI parameters are address placeholders (the `{param}` expressions in a channel address).
* They are string-only and, because every placeholder in the address must be supplied, always
* required. We therefore map them to `path` parameters (the closest OpenAPI analog) and fold the
* `enum`, `default`, and `examples` fields onto a synthetic string schema so the schema renderer
* can display them. The `location` runtime expression has no display equivalent and is dropped.
*/
var adaptAsyncApiParameters = (parameters) => {
	if (!parameters) return [];
	return Object.entries(parameters).map(([name, value]) => {
		const parameter = getResolvedRef(value) ?? {};
		const schema = { type: "string" };
		if (parameter.enum) schema.enum = parameter.enum;
		if (parameter.default !== void 0) schema.default = parameter.default;
		if (parameter.examples) schema.examples = parameter.examples;
		const result = {
			name,
			in: "path",
			required: true,
			schema
		};
		if (parameter.description) result.description = parameter.description;
		return result;
	});
};
//#endregion
export { adaptAsyncApiParameters };

//# sourceMappingURL=adapt-async-api-parameters.js.map