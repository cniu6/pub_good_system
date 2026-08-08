import { useLocalization } from "../../../features/localization/use-localization.js";
import { getCycleKey } from "./helpers/schema-cycle.js";
import { REQUEST_BODY_COMPOSITION_INDEX_SYMBOL } from "../../../features/Operation/request-body-composition-index.js";
import { getSchemaType } from "./helpers/get-schema-type.js";
import { partitionAllOfCompositions } from "./helpers/partition-all-of-compositions.js";
import { getModelNameFromSchema } from "./helpers/schema-name.js";
import Schema_default from "./Schema.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, inject, normalizeClass, openBlock, ref, renderList, resolveComponent, toDisplayString, unref, watch, withCtx } from "vue";
import { ScalarListbox } from "@scalar/components/listbox";
import { ScalarIconCaretDown } from "@scalar/icons";
import { resolve } from "@scalar/workspace-store/resolve";
import { isDefined } from "@scalar/helpers/array/is-defined";
//#region src/components/Content/Schema/SchemaComposition.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "property-rule" };
var _hoisted_2 = {
	class: "composition-selector bg-b-1.5 hover:bg-b-2 flex w-full cursor-pointer items-center gap-1 rounded-t-lg border px-2.5 py-2.5 pr-3 text-left",
	type: "button"
};
var _hoisted_3 = { class: "text-c-2" };
var _hoisted_4 = {
	key: 0,
	class: "text-red"
};
var _hoisted_5 = { class: "composition-panel" };
var SchemaComposition_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaComposition",
	props: {
		composition: {},
		discriminator: {},
		name: {},
		schema: {},
		level: {},
		compact: {
			type: Boolean,
			default: false
		},
		hideHeading: {
			type: Boolean,
			default: false
		},
		hideModelNames: { type: Boolean },
		breadcrumb: {},
		eventBus: {},
		options: {},
		schemaContext: {},
		compositionPath: {}
	},
	setup(__props) {
		const props = __props;
		const { translate } = useLocalization();
		/**
		* Split an `allOf` into an ordered list of segments (object chunks + choice
		* pickers) so multiple mutually-exclusive selections each render their own
		* picker, in the position they were declared, instead of all but the first
		* being dropped and the rest bubbling to the end.
		*/
		const allOfSegments = computed(() => props.composition === "allOf" ? partitionAllOfCompositions(props.schema).segments : []);
		/** The current composition */
		const composition = computed(() => [props.schema[props.composition]].flat().map((schema) => ({
			value: resolve.schema(schema),
			original: schema
		})).filter((it) => isDefined(it.value)));
		/**
		* Generate listbox options for the composition selector.
		* Each option represents a schema in the composition with a human-readable label.
		* Prefers schema title/name over structural type when present.
		*/
		const listboxOptions = computed(() => composition.value.map((schema, index) => {
			const resolved = resolve.schema(schema.original);
			const label = (getModelNameFromSchema(resolved)?.label ?? getSchemaType(resolved)) || translate("schema.schema");
			return {
				id: String(index),
				label
			};
		}));
		const compositionSelectionKey = computed(() => props.compositionPath?.length ? [...props.compositionPath, props.composition].join(".") : "");
		/** When this composition is in the request body, sync selection with the example snippet */
		const requestBodyCompositionSelectionRef = inject(REQUEST_BODY_COMPOSITION_INDEX_SYMBOL, void 0);
		const initialSelectedIndex = computed(() => {
			if (props.schemaContext !== "requestBody" || !requestBodyCompositionSelectionRef?.value || !compositionSelectionKey.value) return 0;
			const selectedIndex = requestBodyCompositionSelectionRef.value[compositionSelectionKey.value];
			if (typeof selectedIndex !== "number" || Number.isNaN(selectedIndex)) return 0;
			return Math.max(0, Math.min(selectedIndex, listboxOptions.value.length - 1));
		});
		/**
		* Two-way computed property for the selected option.
		* Handles conversion between the selected index and the listbox option format.
		*/
		const selectedOption = ref();
		watch([listboxOptions, initialSelectedIndex], ([options, selectedIndex]) => {
			if (!selectedOption.value || !options.some((option) => option.id === selectedOption.value?.id)) selectedOption.value = options[selectedIndex] ?? options[0];
		}, { immediate: true });
		const compositionLabel = (type) => translate(`schema.${type}`);
		/** Inside the currently selected composition */
		const selectedComposition = computed(() => composition.value[Number(selectedOption.value?.id ?? "0")]?.value);
		/**
		* The request body card renders the merged `allOf` description on its outer card
		* (see `Schema.vue`), but only for the top-level request body schema. For that
		* single composition we hide the nested merged `Schema`'s description so the text
		* is not shown twice. Nested request-body compositions (deeper properties) are
		* not shown on the outer card and would otherwise lose their description
		* entirely, because the property row already skips it when `allOf` is present.
		*
		* The top-level request body composition is the one whose `compositionPath` is
		* still the request body root (`['requestBody']`); nested compositions append
		* property segments and therefore have a longer path.
		*/
		const isRequestBodyRootComposition = computed(() => props.schemaContext === "requestBody" && props.compositionPath?.length === 1);
		/**
		* Cycle key for the selected composition member, derived from its raw
		* (unresolved) value so a member that references an ancestor is detected as a
		* cycle.
		*/
		const selectedCompositionCycleKey = computed(() => getCycleKey(composition.value[Number(selectedOption.value?.id ?? "0")]?.original));
		/**
		* Controls whether the nested schema is displayed. When expanding all schema
		* properties we open it by default; the nested Schema handles cycle detection,
		* so finite compositions render fully while recursive ones still stop.
		*/
		const showNestedSchema = ref(!!props.options.expandAllSchemaProperties);
		if (requestBodyCompositionSelectionRef && props.schemaContext === "requestBody" && compositionSelectionKey.value) watch(selectedOption, (option) => {
			const index = option ? Number(option.id) : 0;
			if (!Number.isNaN(index)) requestBodyCompositionSelectionRef.value = {
				...requestBodyCompositionSelectionRef.value,
				[compositionSelectionKey.value]: index
			};
		}, { immediate: true });
		return (_ctx, _cache) => {
			const _component_SchemaComposition = resolveComponent("SchemaComposition", true);
			return openBlock(), createElementBlock("div", _hoisted_1, [props.composition === "allOf" ? (openBlock(true), createElementBlock(Fragment, { key: 0 }, renderList(allOfSegments.value, (segment, segmentIndex) => {
				return openBlock(), createElementBlock(Fragment, { key: segmentIndex }, [segment.kind === "object" ? (openBlock(), createBlock(Schema_default, {
					key: 0,
					breadcrumb: __props.breadcrumb,
					compact: __props.compact,
					compositionPath: __props.compositionPath,
					discriminator: __props.discriminator,
					eventBus: __props.eventBus,
					hideDescription: isRequestBodyRootComposition.value,
					hideHeading: __props.hideHeading,
					hideModelNames: __props.hideModelNames,
					level: __props.level + 1,
					name: __props.name,
					noncollapsible: true,
					options: __props.options,
					schema: segment.schema,
					schemaContext: __props.schemaContext
				}, null, 8, [
					"breadcrumb",
					"compact",
					"compositionPath",
					"discriminator",
					"eventBus",
					"hideDescription",
					"hideHeading",
					"hideModelNames",
					"level",
					"name",
					"options",
					"schema",
					"schemaContext"
				])) : (openBlock(), createBlock(_component_SchemaComposition, {
					key: 1,
					breadcrumb: __props.breadcrumb,
					compact: __props.compact,
					composition: segment.composition,
					compositionPath: [...__props.compositionPath ?? [], String(segment.choiceIndex)],
					eventBus: __props.eventBus,
					hideHeading: __props.hideHeading,
					hideModelNames: __props.hideModelNames,
					level: __props.level,
					options: __props.options,
					schema: segment.value,
					schemaContext: __props.schemaContext
				}, null, 8, [
					"breadcrumb",
					"compact",
					"composition",
					"compositionPath",
					"eventBus",
					"hideHeading",
					"hideModelNames",
					"level",
					"options",
					"schema",
					"schemaContext"
				]))], 64);
			}), 128)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createVNode(unref(ScalarListbox), {
				modelValue: selectedOption.value,
				"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => selectedOption.value = $event),
				options: listboxOptions.value,
				resize: ""
			}, {
				default: withCtx(() => [createElementVNode("button", _hoisted_2, [
					createElementVNode("span", _hoisted_3, toDisplayString(compositionLabel(props.composition)), 1),
					createElementVNode("span", { class: normalizeClass(["composition-selector-label text-c-1", { "line-through": selectedComposition.value?.deprecated }]) }, toDisplayString(selectedOption.value?.label || unref(translate)("schema.schema")), 3),
					selectedComposition.value?.deprecated ? (openBlock(), createElementBlock("div", _hoisted_4, toDisplayString(unref(translate)("common.deprecated")), 1)) : createCommentVNode("", true),
					createVNode(unref(ScalarIconCaretDown))
				])]),
				_: 1
			}, 8, ["modelValue", "options"]), createElementVNode("div", _hoisted_5, [!showNestedSchema.value && __props.level > 2 ? (openBlock(), createElementBlock("button", {
				key: 0,
				class: "bg-b-1 hover:bg-b-2 text-c-1 flex w-full items-center justify-center gap-2 rounded-b-lg border border-t-0 px-2 py-2 text-sm font-medium transition-colors",
				type: "button",
				onClick: _cache[1] || (_cache[1] = ($event) => showNestedSchema.value = true)
			}, [createTextVNode(toDisplayString(unref(translate)("schema.showSchemaDetails")) + " ", 1), createVNode(unref(ScalarIconCaretDown), { class: "h-3 w-3" })])) : (openBlock(), createBlock(Schema_default, {
				key: selectedOption.value?.id ?? "0",
				breadcrumb: __props.breadcrumb,
				compact: __props.compact,
				compositionPath: __props.compositionPath,
				cycleKey: selectedCompositionCycleKey.value,
				discriminator: __props.discriminator,
				eventBus: __props.eventBus,
				hideHeading: __props.hideHeading,
				hideModelNames: __props.hideModelNames,
				level: __props.level + 1,
				name: __props.name,
				noncollapsible: true,
				options: __props.options,
				schema: selectedComposition.value,
				schemaContext: __props.schemaContext
			}, null, 8, [
				"breadcrumb",
				"compact",
				"compositionPath",
				"cycleKey",
				"discriminator",
				"eventBus",
				"hideHeading",
				"hideModelNames",
				"level",
				"name",
				"options",
				"schema",
				"schemaContext"
			]))])], 64))]);
		};
	}
});
//#endregion
export { SchemaComposition_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaComposition.vue.script.js.map