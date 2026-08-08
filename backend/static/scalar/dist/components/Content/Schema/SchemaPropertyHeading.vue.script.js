import { useLocalization } from "../../../features/localization/use-localization.js";
import Badge_default from "../../Badge/Badge.vue.js";
import ScreenReader_default from "../../ScreenReader.vue.js";
import { getSchemaType } from "./helpers/get-schema-type.js";
import { getModelNameFromSchema } from "./helpers/schema-name.js";
import LinkButton_default from "./LinkButton.vue.js";
import RenderString_default from "./RenderString.vue.js";
import SchemaPropertyDefault_default from "./SchemaPropertyDefault.vue.js";
import SchemaPropertyDetail_default from "./SchemaPropertyDetail.vue.js";
import SchemaPropertyExamples_default from "./SchemaPropertyExamples.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createSlots, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, renderList, renderSlot, toDisplayString, toRef, unref, withCtx } from "vue";
import { resolve } from "@scalar/workspace-store/resolve";
import { isDefined } from "@scalar/helpers/array/is-defined";
import { isArraySchema, isNumberSchema, isStringSchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/components/Content/Schema/SchemaPropertyHeading.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "property-heading" };
var _hoisted_2 = {
	key: 1,
	class: "property-discriminator"
};
var _hoisted_3 = {
	key: 3,
	class: "property-additional"
};
var _hoisted_4 = {
	key: 4,
	class: "property-deprecated"
};
var _hoisted_5 = {
	key: 5,
	class: "property-const"
};
var _hoisted_6 = {
	key: 7,
	class: "property-write-only"
};
var _hoisted_7 = {
	key: 8,
	class: "property-read-only"
};
var _hoisted_8 = {
	key: 9,
	class: "property-required"
};
/**
* Constraints that live on a leaf (string or number) schema: length, pattern,
* format and numeric ranges. Extracted so they can be surfaced for a primitive
* schema as well as for a primitive array's items, which are not rendered on
* their own. See https://github.com/scalar/scalar/issues/9690
*/
var SchemaPropertyHeading_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaPropertyHeading",
	props: {
		value: {},
		enum: { type: Boolean },
		isDiscriminator: {
			type: Boolean,
			default: false
		},
		required: {
			type: Boolean,
			default: false
		},
		additional: { type: Boolean },
		withExamples: {
			type: Boolean,
			default: true
		},
		hideModelNames: {
			type: Boolean,
			default: false
		},
		modelName: {},
		propertyNames: {},
		eventBus: { default: null }
	},
	setup(__props) {
		const props = __props;
		const { translate } = useLocalization();
		const valueRef = toRef(props, "value");
		const constValue = computed(() => {
			if (!valueRef.value) return;
			const schema = valueRef.value;
			if (schema.const !== void 0) return schema.const;
			if (schema.enum?.length === 1) return schema.enum[0];
			if (isArraySchema(schema) && schema.items) {
				const items = resolve.schema(schema.items);
				if (isDefined(items.const)) return items.const;
				if (items.enum?.length === 1) return items.enum[0];
			}
		});
		const getLeafConstraints = (schema) => {
			const properties = [];
			if (isStringSchema(schema)) {
				if (schema.minLength) properties.push({
					key: "min-length",
					prefix: `${translate("common.minLength")}: `,
					value: schema.minLength
				});
				if (schema.maxLength) properties.push({
					key: "max-length",
					prefix: `${translate("common.maxLength")}: `,
					value: schema.maxLength
				});
				if (schema.pattern) properties.push({
					key: "pattern",
					value: schema.pattern,
					code: true,
					truncate: true
				});
			}
			if ((isStringSchema(schema) || isNumberSchema(schema)) && schema.format) properties.push({
				key: "format",
				value: schema.format,
				truncate: true
			});
			if (isNumberSchema(schema)) {
				if (isDefined(schema.exclusiveMinimum)) properties.push({
					key: "exclusive-minimum",
					prefix: `${translate("common.greaterThan")}: `,
					value: schema.exclusiveMinimum
				});
				if (isDefined(schema.minimum)) properties.push({
					key: "minimum",
					prefix: `${translate("common.min")}: `,
					value: schema.minimum
				});
				if (isDefined(schema.exclusiveMaximum)) properties.push({
					key: "exclusive-maximum",
					prefix: `${translate("common.lessThan")}: `,
					value: schema.exclusiveMaximum
				});
				if (isDefined(schema.maximum)) properties.push({
					key: "maximum",
					prefix: `${translate("common.max")}: `,
					value: schema.maximum
				});
				if (isDefined(schema.multipleOf)) properties.push({
					key: "multiple-of",
					prefix: `${translate("common.multipleOf")}: `,
					value: schema.multipleOf
				});
			}
			return properties;
		};
		const validationProperties = computed(() => {
			if (!valueRef.value) return [];
			const schema = valueRef.value;
			const properties = [];
			if (isArraySchema(schema)) {
				if (schema.minItems || schema.maxItems) properties.push({
					key: "array-range",
					value: `${schema.minItems || ""}…${schema.maxItems || ""}`
				});
				if (schema.uniqueItems) properties.push({
					key: "unique-items",
					value: `${translate("common.unique")}!`
				});
			}
			properties.push(...getLeafConstraints(schema));
			if (isArraySchema(schema) && schema.items) properties.push(...getLeafConstraints(resolve.schema(schema.items)));
			return properties;
		});
		/** Link data for navigating to the referenced model in the sidebar. */
		const modelLink = computed(() => {
			if (!props.value) return null;
			if (props.hideModelNames) return null;
			if (props.modelName) return {
				schemaKey: props.modelName,
				label: props.modelName
			};
			const modelName = getModelNameFromSchema(props.value);
			if (modelName) return {
				schemaKey: modelName.schemaKey,
				label: modelName.label
			};
			if (isArraySchema(props.value) && props.value.items) {
				const itemName = getModelNameFromSchema(props.value.items);
				return itemName ? {
					schemaKey: itemName.schemaKey,
					label: `${itemName.label}[]`
				} : null;
			}
			return null;
		});
		/** Check if we should show the type information */
		const shouldShowType = computed(() => {
			if (!props.value || !("type" in props.value)) return false;
			if (props.value.type === "array") return true;
			return !constValue.value;
		});
		/** Get the display type */
		const displayType = computed(() => {
			if (!props.value) return "";
			return getSchemaType(props.value);
		});
		/**
		* Type and format of the property keys, derived from the propertyNames schema.
		*
		* For a map keyed by UUIDs this renders e.g. "string · uuid" so the key
		* constraints are not lost. Returns undefined when there is nothing to show.
		*/
		const propertyNamesDetail = computed(() => {
			const schema = props.propertyNames;
			if (!schema) return;
			const parts = [getSchemaType(schema)];
			if ("format" in schema && typeof schema.format === "string") parts.push(schema.format);
			const detail = parts.filter(Boolean).join(" · ");
			return detail.length > 0 ? detail : void 0;
		});
		const exampleValue = computed(() => {
			if (props.value && "example" in props.value && props.value.example !== void 0) return props.value.example;
			if (props.value && isArraySchema(props.value)) {
				const itemsSchema = resolve.schema(props.value.items);
				if (itemsSchema && "example" in itemsSchema && itemsSchema.example !== void 0) return itemsSchema.example;
			}
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", _hoisted_1, [
				_ctx.$slots.name ? (openBlock(), createElementBlock("div", {
					key: 0,
					class: normalizeClass(["property-name", { deprecated: props.value?.deprecated }])
				}, [renderSlot(_ctx.$slots, "name", {}, void 0, true)], 2)) : createCommentVNode("", true),
				props.isDiscriminator ? (openBlock(), createElementBlock("div", _hoisted_2, toDisplayString(unref(translate)("common.discriminator")), 1)) : createCommentVNode("", true),
				props.value ? (openBlock(), createElementBlock(Fragment, { key: 2 }, [
					shouldShowType.value ? (openBlock(), createBlock(SchemaPropertyDetail_default, {
						key: 0,
						truncate: ""
					}, {
						default: withCtx(() => [
							createVNode(ScreenReader_default, null, {
								default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.type")) + ":", 1)]),
								_: 1
							}),
							createTextVNode(" " + toDisplayString(displayType.value) + " ", 1),
							modelLink.value ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [_cache[1] || (_cache[1] = createTextVNode(" · ", -1)), props.eventBus && modelLink.value.schemaKey ? (openBlock(), createBlock(LinkButton_default, {
								key: 0,
								onClick: _cache[0] || (_cache[0] = ($event) => props.eventBus.emit("scroll-to:model-by-name", { name: modelLink.value.schemaKey }))
							}, {
								default: withCtx(() => [createTextVNode(toDisplayString(modelLink.value.label), 1)]),
								_: 1
							})) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(modelLink.value.label), 1)], 64))], 64)) : createCommentVNode("", true)
						]),
						_: 1
					})) : createCommentVNode("", true),
					propertyNamesDetail.value ? (openBlock(), createBlock(SchemaPropertyDetail_default, {
						key: 1,
						truncate: ""
					}, {
						prefix: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.keys")) + ":", 1)]),
						default: withCtx(() => [createTextVNode(" " + toDisplayString(propertyNamesDetail.value), 1)]),
						_: 1
					})) : createCommentVNode("", true),
					(openBlock(true), createElementBlock(Fragment, null, renderList(validationProperties.value, (property) => {
						return openBlock(), createBlock(SchemaPropertyDetail_default, {
							key: property.key,
							code: property.code,
							truncate: property.truncate
						}, createSlots({
							default: withCtx(() => [property.key === "format" ? (openBlock(), createBlock(ScreenReader_default, { key: 0 }, {
								default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.format")) + ": ", 1)]),
								_: 1
							})) : property.key === "pattern" ? (openBlock(), createBlock(ScreenReader_default, { key: 1 }, {
								default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.pattern")) + ": ", 1)]),
								_: 1
							})) : createCommentVNode("", true), createTextVNode(" " + toDisplayString(property.value), 1)]),
							_: 2
						}, [property.prefix ? {
							name: "prefix",
							fn: withCtx(() => [createTextVNode(toDisplayString(property.prefix), 1)]),
							key: "0"
						} : void 0]), 1032, ["code", "truncate"]);
					}), 128)),
					props.enum ? (openBlock(), createBlock(SchemaPropertyDetail_default, { key: 2 }, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.enum")), 1)]),
						_: 1
					})) : createCommentVNode("", true)
				], 64)) : createCommentVNode("", true),
				props.additional ? (openBlock(), createElementBlock("div", _hoisted_3, [props.value?.["x-additionalPropertiesName"] ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createTextVNode(toDisplayString(props.value["x-additionalPropertiesName"]), 1)], 64)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(unref(translate)("common.additionalProperties")), 1)], 64))])) : createCommentVNode("", true),
				props.value?.deprecated ? (openBlock(), createElementBlock("div", _hoisted_4, [createVNode(unref(Badge_default), null, {
					default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.deprecated")), 1)]),
					_: 1
				})])) : createCommentVNode("", true),
				constValue.value !== void 0 ? (openBlock(), createElementBlock("div", _hoisted_5, [createVNode(SchemaPropertyDetail_default, { truncate: "" }, {
					prefix: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.const")) + ": ", 1)]),
					default: withCtx(() => [createVNode(RenderString_default, { value: constValue.value }, null, 8, ["value"])]),
					_: 1
				})])) : (openBlock(), createElementBlock(Fragment, { key: 6 }, [props.value?.nullable === true ? (openBlock(), createBlock(SchemaPropertyDetail_default, { key: 0 }, {
					default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("common.nullable")), 1)]),
					_: 1
				})) : createCommentVNode("", true)], 64)),
				props.value?.writeOnly ? (openBlock(), createElementBlock("div", _hoisted_6, toDisplayString(unref(translate)("common.writeOnly")), 1)) : props.value?.readOnly ? (openBlock(), createElementBlock("div", _hoisted_7, toDisplayString(unref(translate)("common.readOnly")), 1)) : createCommentVNode("", true),
				props.required ? (openBlock(), createElementBlock("div", _hoisted_8, toDisplayString(unref(translate)("common.required")), 1)) : createCommentVNode("", true),
				createVNode(SchemaPropertyDefault_default, { value: props.value?.default }, null, 8, ["value"]),
				props.withExamples ? (openBlock(), createBlock(SchemaPropertyExamples_default, {
					key: 10,
					example: exampleValue.value,
					examples: props.value?.examples
				}, null, 8, ["example", "examples"])) : createCommentVNode("", true)
			]);
		};
	}
});
//#endregion
export { SchemaPropertyHeading_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaPropertyHeading.vue.script.js.map