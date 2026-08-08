import { useLocalization } from "../../../features/localization/use-localization.js";
import SchemaEnumPropertyItem_default from "./SchemaEnumPropertyItem.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, ref, renderList, toDisplayString, unref, withCtx } from "vue";
import { ScalarIconPlus } from "@scalar/icons";
import { ScalarButton } from "@scalar/components/button";
import { resolve } from "@scalar/workspace-store/resolve";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/components/Content/Schema/SchemaEnums.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "property-enum"
};
var _hoisted_2 = {
	key: 0,
	class: "property-enum-property-names"
};
var _hoisted_3 = {
	key: 1,
	class: "property-enum-property-names"
};
var _hoisted_4 = { class: "property-enum-values" };
var _hoisted_5 = { key: 1 };
var ENUM_DISPLAY_THRESHOLD = 9;
var INITIAL_VISIBLE_COUNT = 5;
var THIN_SPACE = " ";
/**
* Resolves the schema that carries the enum values.
* For arrays, the enum and its x-enum-* metadata live on the items schema, so
* both the values and their varnames/descriptions have to be read from there.
*/
var SchemaEnums_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaEnums",
	props: {
		value: {},
		propertyNames: { type: Boolean }
	},
	setup(__props) {
		const { translate } = useLocalization();
		const enumSchema = computed(() => {
			if (!__props.value) return;
			if (__props.value.enum) return __props.value;
			return isArraySchema(__props.value) ? resolve.schema(__props.value.items) : void 0;
		});
		/**
		* Extracts enum values from the schema object.
		* Handles both direct enum values and nested enum arrays.
		*/
		const enumValues = computed(() => enumSchema.value?.enum ?? []);
		/**
		* Determines if we should show the long enum list UI.
		* When there are many enum values, we initially show only a subset.
		*/
		const shouldUseLongListDisplay = computed(() => enumValues.value.length > ENUM_DISPLAY_THRESHOLD);
		const initialVisibleCount = computed(() => shouldUseLongListDisplay.value ? INITIAL_VISIBLE_COUNT : enumValues.value.length);
		const visibleEnumValues = computed(() => enumValues.value.slice(0, initialVisibleCount.value));
		const hiddenEnumValues = computed(() => enumValues.value.slice(initialVisibleCount.value));
		/**
		* Gets the description for an enum value.
		* Supports both array and object formats for x-enumDescriptions.
		*/
		const getEnumValueDescription = (enumValue, index) => {
			const schema = enumSchema.value;
			const descriptions = schema?.["x-enumDescriptions"] ?? schema?.["x-enum-descriptions"];
			if (!descriptions) return;
			if (Array.isArray(descriptions)) return descriptions[index];
			if (typeof descriptions === "object" && descriptions !== null) return descriptions[String(enumValue)];
		};
		/**
		* Formats an enum value with its variable name if available.
		* This supports both x-enum-varnames and x-enumNames extensions.
		*/
		const formatEnumValueWithName = (enumValue, index) => {
			const varNames = enumSchema.value?.["x-enum-varnames"] ?? enumSchema.value?.["x-enumNames"];
			const varName = Array.isArray(varNames) ? varNames[index] : void 0;
			return varName ? `${enumValue}${THIN_SPACE}=${THIN_SPACE}${varName}` : String(enumValue);
		};
		/**
		* Controls whether the hidden enum values are visible.
		*/
		const isExpanded = ref(false);
		const toggleExpanded = () => {
			isExpanded.value = !isExpanded.value;
		};
		return (_ctx, _cache) => {
			return enumValues.value.length > 0 ? (openBlock(), createElementBlock("div", _hoisted_1, [__props.propertyNames ? (openBlock(), createElementBlock("div", _hoisted_2, toDisplayString(unref(translate)("common.propertyNames")), 1)) : (openBlock(), createElementBlock("div", _hoisted_3, toDisplayString(unref(translate)("common.values")), 1)), createElementVNode("ul", _hoisted_4, [
				(openBlock(true), createElementBlock(Fragment, null, renderList(visibleEnumValues.value, (enumValue, index) => {
					return openBlock(), createBlock(SchemaEnumPropertyItem_default, {
						key: String(enumValue),
						description: getEnumValueDescription(enumValue, index),
						label: formatEnumValueWithName(enumValue, index)
					}, null, 8, ["description", "label"]);
				}), 128)),
				shouldUseLongListDisplay.value && isExpanded.value ? (openBlock(true), createElementBlock(Fragment, { key: 0 }, renderList(hiddenEnumValues.value, (enumValue, index) => {
					return openBlock(), createBlock(SchemaEnumPropertyItem_default, {
						key: String(enumValue),
						description: getEnumValueDescription(enumValue, initialVisibleCount.value + index),
						label: formatEnumValueWithName(enumValue, initialVisibleCount.value + index)
					}, null, 8, ["description", "label"]);
				}), 128)) : createCommentVNode("", true),
				shouldUseLongListDisplay.value ? (openBlock(), createElementBlock("li", _hoisted_5, [createVNode(unref(ScalarButton), {
					class: "enum-toggle-button my-2 flex h-fit gap-1 rounded-full border py-1.5 pr-2.5 pl-2 leading-none",
					variant: "ghost",
					onClick: toggleExpanded
				}, {
					default: withCtx(() => [createVNode(unref(ScalarIconPlus), {
						class: normalizeClass({ "rotate-45": isExpanded.value }),
						weight: "bold"
					}, null, 8, ["class"]), createTextVNode(" " + toDisplayString(isExpanded.value ? unref(translate)("common.hideValues") : unref(translate)("common.showAllValues")), 1)]),
					_: 1
				})])) : createCommentVNode("", true)
			])])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SchemaEnums_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaEnums.vue.script.js.map