import { isTypeObject } from "./is-type-object.js";
import { resolve } from "@scalar/workspace-store/resolve";
//#region src/components/Content/Schema/helpers/sort-property-names.ts
/** Take a list of property names and reduce it back into an object */
var reduceNamesToObject = (names, properties) => names.reduce((acc, name) => {
	const prop = properties?.[name];
	if (prop) acc[name] = prop;
	return acc;
}, {});
/** Sort property names in an object schema */
var sortPropertyNames = (schema, discriminator, { hideReadOnly = false, hideWriteOnly = false, orderSchemaPropertiesBy = "alpha", orderRequiredPropertiesFirst = true } = {}) => {
	if (!isTypeObject(schema) || !schema.properties) return [];
	const propertyNames = Object.keys(schema.properties);
	const requiredPropertiesSet = new Set(schema.required || []);
	return propertyNames.sort((a, b) => {
		const aDiscriminator = a === discriminator?.propertyName;
		const bDiscriminator = b === discriminator?.propertyName;
		const aRequired = requiredPropertiesSet.has(a);
		const bRequired = requiredPropertiesSet.has(b);
		if (aDiscriminator && !bDiscriminator) return -1;
		if (!aDiscriminator && bDiscriminator) return 1;
		const aSchema = schema.properties?.[a];
		const bSchema = schema.properties?.[b];
		const aOrder = aSchema && typeof aSchema === "object" && "x-order" in aSchema ? aSchema["x-order"] : void 0;
		const bOrder = bSchema && typeof bSchema === "object" && "x-order" in bSchema ? bSchema["x-order"] : void 0;
		if (aOrder !== void 0 && bOrder !== void 0) return Number(aOrder) - Number(bOrder);
		if (aOrder !== void 0 && bOrder === void 0) return -1;
		if (aOrder === void 0 && bOrder !== void 0) return 1;
		if (orderRequiredPropertiesFirst) {
			if (aRequired && !bRequired) return -1;
			if (!aRequired && bRequired) return 1;
		}
		if (orderSchemaPropertiesBy === "alpha") return a.localeCompare(b);
		return 0;
	}).filter((property) => {
		const resolved = schema.properties && resolve.schema(schema.properties[property]);
		if (hideReadOnly && resolved?.readOnly === true) return false;
		if (hideWriteOnly && resolved?.writeOnly === true) return false;
		return true;
	});
};
//#endregion
export { reduceNamesToObject, sortPropertyNames };

//# sourceMappingURL=sort-property-names.js.map