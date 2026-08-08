import WithBreadcrumb_default from "../../Anchor/WithBreadcrumb.vue.js";
import SpecificationExtension_default from "../../../features/specification-extension/SpecificationExtension.vue.js";
import { resolveDynamicSchema, useDynamicScope } from "./helpers/dynamic-scope.js";
import { hasComplexArrayItems } from "./helpers/has-complex-array-items.js";
import { getCompositionsToRender } from "./helpers/get-compositions-to-render.js";
import { isTypeObject } from "./helpers/is-type-object.js";
import { getCycleKey } from "./helpers/schema-cycle.js";
import SchemaComposition_default from "./SchemaComposition.vue.js";
import { getEnumValues } from "./helpers/get-enum-values.js";
import { getPropertyDescription } from "./helpers/get-property-description.js";
import { optimizeValueForDisplay } from "./helpers/optimize-value-for-display.js";
import { shouldDisplayDescription } from "./helpers/should-display-description.js";
import { shouldDisplayHeading } from "./helpers/should-display-heading.js";
import SchemaEnums_default from "./SchemaEnums.vue.js";
import SchemaPropertyHeading_default from "./SchemaPropertyHeading.vue.js";
import Schema_default from "./Schema.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createSlots, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, renderList, resolveDynamicComponent, toDisplayString, unref, withCtx } from "vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
import { resolve } from "@scalar/workspace-store/resolve";
import { isDynamicRef, resolveDynamicRef } from "@scalar/workspace-store/helpers/dynamic-ref";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
import { ScalarWrappingText } from "@scalar/components/wrapping-text";
//#region src/components/Content/Schema/SchemaProperty.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "property-name-pattern-properties"
};
var _hoisted_2 = {
	key: 1,
	class: "property-name-additional-properties"
};
var _hoisted_3 = {
	key: 1,
	class: "property-description"
};
var _hoisted_4 = {
	key: 4,
	class: "children"
};
var _hoisted_5 = {
	key: 5,
	class: "children"
};
/**
* Note: We're taking in a prop called `value` which should be a JSON Schema.
*
* We're using `optimizeValueForDisplay` to merge null types in compositions (anyOf, allOf, oneOf, not).
* So you should basically use the optimizedValue everywhere in the component.
*/
/** Composition keywords that hold a list of schemas and can be flattened when they contain a single member. */
var SchemaProperty_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaProperty",
	props: {
		is: {},
		schema: {},
		noncollapsible: { type: Boolean },
		level: { default: 0 },
		name: {},
		required: {
			type: Boolean,
			default: false
		},
		compact: {
			type: Boolean,
			default: false
		},
		discriminator: {},
		description: {},
		hideModelNames: {
			type: Boolean,
			default: false
		},
		hideHeading: { type: Boolean },
		modelName: {},
		variant: {},
		breadcrumb: {},
		eventBus: {},
		options: {},
		propertyNamesEnum: {},
		propertyNamesSchema: {},
		schemaContext: {},
		compositionPath: {},
		compositionPathSegment: {},
		cycleKey: {}
	},
	setup(__props) {
		const SINGLE_ITEM_COMPOSITIONS = [
			"oneOf",
			"anyOf",
			"allOf"
		];
		const props = __props;
		/** The dynamic scope inherited from the enclosing schema resources, used to bind `$dynamicRef`s. */
		const dynamicScope = useDynamicScope();
		/**
		* Simplified composition with `null` type.
		*
		* A top-level `$dynamicRef` (e.g. a linked-list `next` node) is bound to its concrete type via the
		* dynamic scope first; for ordinary schemas this is a no-op.
		*/
		const optimizedValue = computed(() => optimizeValueForDisplay(resolveDynamicSchema(props.schema, dynamicScope)));
		const childBreadcrumb = computed(() => props.breadcrumb ? props.name ? [...props.breadcrumb, props.name] : props.breadcrumb : void 0);
		const currentCompositionPath = computed(() => props.compositionPathSegment ? [...props.compositionPath ?? [], props.compositionPathSegment] : props.compositionPath ?? []);
		const arrayItemsCompositionPath = computed(() => [...currentCompositionPath.value, "items"]);
		const shouldHaveLink = computed(() => props.level <= 2);
		/**
		* The array schema used for item inspection, with a `$dynamicRef` item bound to its concrete type.
		*
		* Returns the schema unchanged unless `items` is a `$dynamicRef` that resolves against the dynamic
		* scope, so ordinary arrays (including `$ref` items) keep their existing behavior exactly.
		*/
		const arrayValueWithBoundItems = computed(() => {
			const value = optimizedValue.value;
			if (!value || !isArraySchema(value) || !isDynamicRef(value.items)) return value;
			const bound = resolveDynamicRef(value.items.$dynamicRef, dynamicScope);
			return bound ? {
				...value,
				items: bound
			} : value;
		});
		/** Checks if array items have complex structure */
		const hasComplexArrayItemsComputed = computed(() => hasComplexArrayItems(arrayValueWithBoundItems.value));
		/** Check if enum should be displayed (from value schema or from propertyNames) */
		const hasEnum = computed(() => enumValues.value.length > 0);
		/** Determine if object properties should be displayed */
		const shouldRenderObjectProperties = computed(() => {
			const value = optimizedValue.value;
			if (!value) return false;
			if (!("properties" in value || "additionalProperties" in value)) return false;
			if ("allOf" in value) return false;
			const type = value.type;
			const isExplicitNonObject = typeof type === "string" && type !== "object";
			return isTypeObject(value) || !isExplicitNonObject;
		});
		/** Determine if array of objects should be rendered */
		const shouldRenderArrayOfObjects = computed(() => {
			const value = optimizedValue.value;
			if (!value || !isArraySchema(value) || typeof value.items !== "object") return false;
			return hasComplexArrayItemsComputed.value;
		});
		/** Extract enum values from schema or array items */
		const enumValues = computed(() => getEnumValues(optimizedValue.value));
		/** Generate property description from type/format */
		const propertyDescription = computed(() => getPropertyDescription(optimizedValue.value));
		/** Determine if description should be displayed */
		const displayDescription = computed(() => shouldDisplayDescription(optimizedValue.value, props.description));
		/**
		* The schema used to render the object's own properties.
		*
		* Composition keywords are stripped so the nested object renders only its
		* properties. The compositions are rendered separately below; leaving them here
		* would route the nested `Schema` back through `SchemaProperty` and recurse.
		*
		* The `discriminator` is stripped too: its variant selector is part of the
		* composition rendered below (including the one inferred from a bare
		* `discriminator.mapping`), so keeping it here would make the nested `Schema`
		* infer and render a second, identical selector.
		*
		* When the property already renders the description, we also drop it to avoid
		* repeating it in the nested object schema card.
		*/
		const objectSchemaForChildren = computed(() => {
			const value = optimizedValue.value;
			if (!value) return value;
			const { oneOf: _oneOf, anyOf: _anyOf, allOf: _allOf, not: _not, discriminator: _discriminator, ...objectSchema } = value;
			if (displayDescription.value && "description" in objectSchema) {
				const { description: _description, ...schemaWithoutDescription } = objectSchema;
				return schemaWithoutDescription;
			}
			return objectSchema;
		});
		/** Determine if property heading should be displayed */
		const shouldDisplayHeadingComputed = computed(() => shouldDisplayHeading(optimizedValue.value, props.name, props.required));
		/** Computes which compositions should be rendered and with which values */
		const compositionsToRender = computed(() => getCompositionsToRender(optimizedValue.value, props.options.document));
		const getCompositionDiscriminator = (composition) => composition === "allOf" ? props.schema?.discriminator ?? props.discriminator : props.schema?.discriminator;
		/**
		* Get resolved array items for rendering (with any `$dynamicRef` bound to the concrete type).
		*
		* When the items are wrapped in a single-item composition (e.g. `items: { allOf: [{ type: 'object', ... }] }`), we
		* flatten that wrapper into its plain form. A composition with a single member is equivalent to that member, so
		* keeping the composition keyword only makes the items render through an extra schema layer, which adds an
		* unnecessary level of nesting and duplicates the item description. See https://github.com/scalar/scalar/issues/5900
		*/
		const resolvedArrayItems = computed(() => {
			const value = arrayValueWithBoundItems.value;
			if (!value || !isArraySchema(value) || typeof value.items !== "object") return;
			const items = resolve.schema(value.items);
			return SINGLE_ITEM_COMPOSITIONS.some((keyword) => Array.isArray(items?.[keyword]) && items[keyword]?.length === 1) ? optimizeValueForDisplay(items) : items;
		});
		/**
		* Cycle key for the array items schema, derived from the raw (unresolved) items
		* so a self-referential array element is detected as a cycle.
		*/
		const arrayItemsCycleKey = computed(() => {
			const value = optimizedValue.value;
			if (!value || !isArraySchema(value)) return;
			return getCycleKey(value.items);
		});
		/** Check if discriminator matches current property */
		const isDiscriminatorProperty = computed(() => Boolean(props.name && props.discriminator?.propertyName === props.name));
		return (_ctx, _cache) => {
			return openBlock(), createBlock(resolveDynamicComponent(__props.is ?? "li"), { class: normalizeClass(["property", [`property--level-${__props.level}`, {
				"property--compact": __props.compact,
				"property--deprecated": optimizedValue.value?.deprecated
			}]]) }, {
				default: withCtx(() => [
					shouldDisplayHeadingComputed.value ? (openBlock(), createBlock(SchemaPropertyHeading_default, {
						key: 0,
						class: "group",
						enum: hasEnum.value,
						eventBus: __props.eventBus,
						hideModelNames: __props.hideModelNames,
						isDiscriminator: isDiscriminatorProperty.value,
						modelName: __props.modelName,
						propertyNames: __props.propertyNamesSchema,
						required: __props.required,
						value: optimizedValue.value
					}, createSlots({ _: 2 }, [__props.name ? {
						name: "name",
						fn: withCtx(() => [createVNode(unref(WithBreadcrumb_default), {
							breadcrumb: shouldHaveLink.value ? childBreadcrumb.value : void 0,
							eventBus: __props.eventBus
						}, {
							default: withCtx(() => [__props.variant === "patternProperties" ? (openBlock(), createElementBlock("span", _hoisted_1, [createVNode(unref(ScalarWrappingText), {
								preset: "property",
								text: __props.name
							}, null, 8, ["text"])])) : __props.variant === "additionalProperties" ? (openBlock(), createElementBlock("span", _hoisted_2, [createVNode(unref(ScalarWrappingText), {
								preset: "property",
								text: __props.name
							}, null, 8, ["text"])])) : (openBlock(), createBlock(unref(ScalarWrappingText), {
								key: 2,
								preset: "property",
								text: __props.name
							}, null, 8, ["text"]))]),
							_: 1
						}, 8, ["breadcrumb", "eventBus"])]),
						key: "0"
					} : void 0, optimizedValue.value?.example !== void 0 ? {
						name: "example",
						fn: withCtx(() => [createTextVNode(" Example: " + toDisplayString(optimizedValue.value.example), 1)]),
						key: "1"
					} : void 0]), 1032, [
						"enum",
						"eventBus",
						"hideModelNames",
						"isDiscriminator",
						"modelName",
						"propertyNames",
						"required",
						"value"
					])) : createCommentVNode("", true),
					displayDescription.value || propertyDescription.value ? (openBlock(), createElementBlock("div", _hoisted_3, [createVNode(unref(ScalarMarkdown), { value: displayDescription.value || propertyDescription.value || "" }, null, 8, ["value"])])) : createCommentVNode("", true),
					__props.propertyNamesEnum && __props.propertyNamesEnum.length > 0 ? (openBlock(), createBlock(SchemaEnums_default, {
						key: 2,
						propertyNames: "",
						value: { enum: __props.propertyNamesEnum }
					}, null, 8, ["value"])) : createCommentVNode("", true),
					enumValues.value.length > 0 && !shouldRenderArrayOfObjects.value ? (openBlock(), createBlock(SchemaEnums_default, {
						key: 3,
						value: optimizedValue.value
					}, null, 8, ["value"])) : createCommentVNode("", true),
					shouldRenderObjectProperties.value ? (openBlock(), createElementBlock("div", _hoisted_4, [createVNode(Schema_default, {
						breadcrumb: childBreadcrumb.value,
						compact: __props.compact,
						compositionPath: currentCompositionPath.value,
						cycleKey: __props.cycleKey,
						eventBus: __props.eventBus,
						hideModelNames: __props.hideModelNames,
						level: __props.level + 1,
						name: __props.name,
						noncollapsible: __props.noncollapsible,
						options: __props.options,
						schema: objectSchemaForChildren.value,
						schemaContext: __props.schemaContext
					}, null, 8, [
						"breadcrumb",
						"compact",
						"compositionPath",
						"cycleKey",
						"eventBus",
						"hideModelNames",
						"level",
						"name",
						"noncollapsible",
						"options",
						"schema",
						"schemaContext"
					])])) : createCommentVNode("", true),
					shouldRenderArrayOfObjects.value && resolvedArrayItems.value ? (openBlock(), createElementBlock("div", _hoisted_5, [createVNode(Schema_default, {
						compact: __props.compact,
						compositionPath: arrayItemsCompositionPath.value,
						cycleKey: arrayItemsCycleKey.value,
						eventBus: __props.eventBus,
						hideModelNames: __props.hideModelNames,
						level: __props.level + 1,
						name: __props.name,
						noncollapsible: __props.noncollapsible,
						options: __props.options,
						schema: unref(resolve).schema(resolvedArrayItems.value),
						schemaContext: __props.schemaContext
					}, null, 8, [
						"compact",
						"compositionPath",
						"cycleKey",
						"eventBus",
						"hideModelNames",
						"level",
						"name",
						"noncollapsible",
						"options",
						"schema",
						"schemaContext"
					])])) : createCommentVNode("", true),
					(openBlock(true), createElementBlock(Fragment, null, renderList(compositionsToRender.value, (compositionData) => {
						return openBlock(), createBlock(SchemaComposition_default, {
							key: compositionData.composition,
							breadcrumb: __props.breadcrumb,
							compact: __props.compact,
							composition: compositionData.composition,
							compositionPath: currentCompositionPath.value,
							discriminator: getCompositionDiscriminator(compositionData.composition),
							eventBus: __props.eventBus,
							hideHeading: __props.hideHeading,
							hideModelNames: __props.hideModelNames,
							level: __props.level,
							name: __props.name,
							noncollapsible: __props.noncollapsible,
							options: __props.options,
							schema: compositionData.value,
							schemaContext: __props.schemaContext
						}, null, 8, [
							"breadcrumb",
							"compact",
							"composition",
							"compositionPath",
							"discriminator",
							"eventBus",
							"hideHeading",
							"hideModelNames",
							"level",
							"name",
							"noncollapsible",
							"options",
							"schema",
							"schemaContext"
						]);
					}), 128)),
					createVNode(unref(SpecificationExtension_default), { value: optimizedValue.value }, null, 8, ["value"])
				]),
				_: 1
			}, 8, ["class"]);
		};
	}
});
//#endregion
export { SchemaProperty_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaProperty.vue.script.js.map