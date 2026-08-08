import { isTypeObject } from "./helpers/is-type-object.js";
import { getCycleKey } from "./helpers/schema-cycle.js";
import { sortPropertyNames } from "./helpers/sort-property-names.js";
import SchemaProperty_default from "./SchemaProperty.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, defineComponent, openBlock, renderList, unref } from "vue";
import { resolve } from "@scalar/workspace-store/resolve";
//#region src/components/Content/Schema/SchemaObjectProperties.vue?vue&type=script&setup=true&lang.ts
var SchemaObjectProperties_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaObjectProperties",
	props: {
		schema: {},
		discriminator: {},
		compact: { type: Boolean },
		hideHeading: { type: Boolean },
		level: {},
		hideModelNames: { type: Boolean },
		breadcrumb: {},
		eventBus: {},
		options: {},
		schemaContext: {},
		compositionPath: {}
	},
	setup(__props) {
		/**
		* Sorts properties by required status first, then alphabetically.
		* Required properties appear first, followed by optional properties.
		*/
		const sortedProperties = computed(() => sortPropertyNames(__props.schema, __props.discriminator, __props.options));
		/**
		* Get the display name for additional properties.
		*
		* Checks x-additionalPropertiesName extension first, then falls back to the
		* propertyNames schema title if available.
		*/
		const getAdditionalPropertiesName = (_additionalProperties, _propertyNames) => {
			const additionalProperties = typeof _additionalProperties === "boolean" ? _additionalProperties : resolve.schema(_additionalProperties);
			if (typeof additionalProperties === "object" && typeof additionalProperties["x-additionalPropertiesName"] === "string" && additionalProperties["x-additionalPropertiesName"].trim().length > 0) return `${additionalProperties["x-additionalPropertiesName"].trim()}`;
			if (_propertyNames) {
				const resolved = resolve.schema(_propertyNames);
				if (resolved?.title) return resolved.title;
			}
			return "propertyName";
		};
		/**
		* Extract enum values from the propertyNames schema.
		*
		* JSON Schema's propertyNames keyword constrains which keys are valid
		* in an object with additionalProperties. When it contains an enum,
		* these are the allowed key names.
		*/
		const getPropertyNamesEnum = (_propertyNames) => {
			if (!_propertyNames) return;
			const resolved = resolve.schema(_propertyNames);
			if (resolved && "enum" in resolved && Array.isArray(resolved.enum) && resolved.enum.length > 0) return resolved.enum;
		};
		/** Enum values for the property keys, derived from propertyNames if present. */
		const additionalPropertiesEnum = computed(() => {
			if (!isTypeObject(__props.schema) || !__props.schema.additionalProperties) return;
			return getPropertyNamesEnum(__props.schema.propertyNames);
		});
		/**
		* The resolved propertyNames schema for the property keys.
		*
		* Surfaces key constraints such as `format` (for example `uuid`) so they are
		* not lost when rendering a map of additional properties.
		*/
		const additionalPropertiesKeySchema = computed(() => {
			if (!isTypeObject(__props.schema) || !__props.schema.additionalProperties) return;
			return __props.schema.propertyNames ? resolve.schema(__props.schema.propertyNames) : void 0;
		});
		/**
		* Keep sibling property descriptions separate from the referenced schema.
		*
		* This allows us to render both:
		* - the property-specific description written next to the $ref, and
		* - the referenced schema's own description (for example discriminator parent docs)
		*/
		const getPropertySchema = (property) => {
			if (!property) return;
			return resolve.schema(property);
		};
		const getPropertyDescription = (property) => {
			if (!property) return;
			return typeof property.description === "string" ? property.description : void 0;
		};
		/**
		* Get the value for additional properties.
		*
		* When additionalProperties is true or an empty object, it should render as { type: 'anything' }.
		* $ref values are resolved before the type check so the referenced schema is rendered correctly.
		*/
		const getAdditionalPropertiesValue = (additionalProperties) => {
			const resolved = typeof additionalProperties === "boolean" ? additionalProperties : resolve.schema(additionalProperties);
			if (resolved === true || typeof resolved === "object" && Object.keys(resolved).length === 0 || typeof resolved !== "object" || !("type" in resolved)) return {
				type: "anything",
				...typeof resolved === "object" ? resolved : {}
			};
			return resolved;
		};
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [
				unref(isTypeObject)(__props.schema) && __props.schema.properties ? (openBlock(true), createElementBlock(Fragment, { key: 0 }, renderList(sortedProperties.value, (property) => {
					return openBlock(), createBlock(SchemaProperty_default, {
						key: property,
						breadcrumb: __props.breadcrumb,
						compact: __props.compact,
						compositionPath: __props.compositionPath,
						compositionPathSegment: property,
						cycleKey: unref(getCycleKey)(__props.schema.properties[property]),
						description: getPropertyDescription(__props.schema.properties[property]),
						discriminator: __props.discriminator,
						eventBus: __props.eventBus,
						hideHeading: __props.hideHeading,
						hideModelNames: __props.hideModelNames,
						level: __props.level,
						name: property,
						options: __props.options,
						required: __props.schema.required?.includes(property),
						schema: getPropertySchema(__props.schema.properties[property]),
						schemaContext: __props.schemaContext
					}, null, 8, [
						"breadcrumb",
						"compact",
						"compositionPath",
						"compositionPathSegment",
						"cycleKey",
						"description",
						"discriminator",
						"eventBus",
						"hideHeading",
						"hideModelNames",
						"level",
						"name",
						"options",
						"required",
						"schema",
						"schemaContext"
					]);
				}), 128)) : createCommentVNode("", true),
				unref(isTypeObject)(__props.schema) && __props.schema.patternProperties ? (openBlock(true), createElementBlock(Fragment, { key: 1 }, renderList(Object.entries(__props.schema.patternProperties), ([key, property]) => {
					return openBlock(), createBlock(SchemaProperty_default, {
						key,
						breadcrumb: __props.breadcrumb,
						compact: __props.compact,
						compositionPath: __props.compositionPath,
						compositionPathSegment: key,
						cycleKey: unref(getCycleKey)(property),
						description: getPropertyDescription(property),
						discriminator: __props.discriminator,
						eventBus: __props.eventBus,
						hideHeading: __props.hideHeading,
						hideModelNames: __props.hideModelNames,
						level: __props.level,
						name: key,
						options: __props.options,
						schema: getPropertySchema(property),
						schemaContext: __props.schemaContext
					}, null, 8, [
						"breadcrumb",
						"compact",
						"compositionPath",
						"compositionPathSegment",
						"cycleKey",
						"description",
						"discriminator",
						"eventBus",
						"hideHeading",
						"hideModelNames",
						"level",
						"name",
						"options",
						"schema",
						"schemaContext"
					]);
				}), 128)) : createCommentVNode("", true),
				unref(isTypeObject)(__props.schema) && __props.schema.additionalProperties ? (openBlock(), createBlock(SchemaProperty_default, {
					key: 2,
					breadcrumb: __props.breadcrumb,
					compact: __props.compact,
					compositionPath: __props.compositionPath,
					compositionPathSegment: getAdditionalPropertiesName(__props.schema.additionalProperties, __props.schema.propertyNames),
					cycleKey: unref(getCycleKey)(__props.schema.additionalProperties),
					discriminator: __props.discriminator,
					eventBus: __props.eventBus,
					hideHeading: __props.hideHeading,
					hideModelNames: __props.hideModelNames,
					level: __props.level,
					name: getAdditionalPropertiesName(__props.schema.additionalProperties, __props.schema.propertyNames),
					noncollapsible: "",
					options: __props.options,
					propertyNamesEnum: additionalPropertiesEnum.value,
					propertyNamesSchema: additionalPropertiesKeySchema.value,
					schema: getAdditionalPropertiesValue(__props.schema.additionalProperties),
					schemaContext: __props.schemaContext,
					variant: "additionalProperties"
				}, null, 8, [
					"breadcrumb",
					"compact",
					"compositionPath",
					"compositionPathSegment",
					"cycleKey",
					"discriminator",
					"eventBus",
					"hideHeading",
					"hideModelNames",
					"level",
					"name",
					"options",
					"propertyNamesEnum",
					"propertyNamesSchema",
					"schema",
					"schemaContext"
				])) : createCommentVNode("", true)
			], 64);
		};
	}
});
//#endregion
export { SchemaObjectProperties_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaObjectProperties.vue.script.js.map