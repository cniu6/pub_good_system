import { XScalarStability } from "@scalar/types/legacy";
//#region src/features/Operation/helpers/operation-stability.ts
/**
* Returns true if an operation is considered deprecated.
*/
var isOperationDeprecated = (operation) => operation.deprecated || operation["x-scalar-stability"] === XScalarStability.Deprecated;
/**
* Get operation stability from deprecated or x-scalar-stability
*/
var getOperationStability = (operation) => operation.deprecated ? XScalarStability.Deprecated : operation["x-scalar-stability"];
/**
* Get Operation stability tailwind color class
*/
var getOperationStabilityColor = (operation) => {
	switch (getOperationStability(operation)) {
		case XScalarStability.Deprecated: return "text-red";
		case XScalarStability.Experimental: return "text-orange";
		case XScalarStability.Stable: return "text-green";
		default: return "";
	}
};
//#endregion
export { getOperationStability, getOperationStabilityColor, isOperationDeprecated };

//# sourceMappingURL=operation-stability.js.map