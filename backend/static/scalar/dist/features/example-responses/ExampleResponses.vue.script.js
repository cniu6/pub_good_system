import { useLocalization } from "../localization/use-localization.js";
import ScreenReader_default from "../../components/ScreenReader.vue.js";
import ExampleSchema_default from "./ExampleSchema.vue.js";
import ExampleResponse_default from "./ExampleResponse.vue.js";
import ExampleResponseTab_default from "./ExampleResponseTab.vue.js";
import ExampleResponseTabList_default from "./ExampleResponseTabList.vue.js";
import { normalizeMimeTypeObject } from "./helpers/normalize-mime-type-object.js";
import { hasResponseContent } from "./helpers/has-response-content.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, ref, renderList, toDisplayString, toValue, unref, useId, vModelCheckbox, watch, withCtx, withDirectives } from "vue";
import { useClipboard } from "@scalar/use-hooks/useClipboard";
import { getExample } from "@scalar/workspace-store/request-example";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { ExamplePicker } from "@scalar/blocks/code-example";
import { ScalarMarkdown } from "@scalar/components/markdown";
import { ScalarIcon } from "@scalar/components/icon";
import { objectKeys } from "@scalar/helpers/object/object-keys";
import { ScalarCard, ScalarCardFooter, ScalarCardSection } from "@scalar/components/card";
//#region src/features/example-responses/ExampleResponses.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 1,
	class: "scalar-card-checkbox"
};
var _hoisted_2 = ["aria-controls"];
var _hoisted_3 = { class: "response-description" };
/**
* TODO: copyToClipboard isn't using the right content if there are multiple examples
*/
var ExampleResponses_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ExampleResponses",
	props: {
		responses: {},
		selectedExample: {},
		eventBus: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		const id = useId();
		const { copyToClipboard } = useClipboard();
		const orderedStatusCodes = computed(() => Object.keys(__props.responses ?? {}).sort());
		const statusCodesWithContent = computed(() => orderedStatusCodes.value.filter((statusCode) => hasResponseContent(getResolvedRef(__props.responses?.[statusCode]), statusCode)));
		const selectedResponseIndex = ref(0);
		/**
		* Clamp the selected index when the filtered list shrinks.
		* Without this, the index can become out of bounds and cause a mismatch
		* between the visible tabs and the displayed content.
		*
		* We re-resolve `selectedExampleKey` the same way `changeTab` does, so the picker keeps the
		* document-wide selection when the newly active response defines it instead of being blanked out.
		*/
		watch(statusCodesWithContent, (codes) => {
			if (codes.length === 0) selectedResponseIndex.value = 0;
			else if (selectedResponseIndex.value >= codes.length) selectedResponseIndex.value = codes.length - 1;
			else return;
			selectedExampleKey.value = resolveExampleKey(__props.selectedExample);
		});
		const currentResponse = computed(() => {
			const currentStatusCode = toValue(statusCodesWithContent)[toValue(selectedResponseIndex)] ?? "";
			return getResolvedRef(__props.responses?.[currentStatusCode]);
		});
		const currentResponseContent = computed(() => {
			const normalizedContent = normalizeMimeTypeObject(currentResponse.value?.content);
			/** All the keys of the normalized content */
			const keys = objectKeys(normalizedContent ?? {});
			return normalizedContent?.[keys[0] ?? ""];
		});
		const hasMultipleExamples = computed(() => !!currentResponseContent.value?.examples && Object.keys(currentResponseContent.value?.examples ?? {}).length > 1);
		const selectedExampleKey = ref("");
		/** Resolve the example key to show, preferring the document-wide selection when this response has it */
		const resolveExampleKey = (preferred) => {
			const keys = Object.keys(currentResponseContent.value?.examples ?? {});
			if (preferred && keys.includes(preferred)) return preferred;
			if (selectedExampleKey.value && keys.includes(selectedExampleKey.value)) return selectedExampleKey.value;
			return keys[0] ?? "";
		};
		selectedExampleKey.value = resolveExampleKey(__props.selectedExample);
		watch(() => __props.selectedExample, (preferred) => {
			selectedExampleKey.value = resolveExampleKey(preferred);
		});
		/** Select an example and sync the choice across the document so other operations follow */
		const selectExample = (key) => {
			selectedExampleKey.value = key;
			__props.eventBus?.emit("workspace:update:selected-example", key);
		};
		/** Get the current example to display */
		const currentExample = computed(() => {
			if (!currentResponseContent.value) return;
			if (hasMultipleExamples.value && selectedExampleKey.value) return currentResponseContent.value.examples?.[selectedExampleKey.value];
			return getExample(currentResponseContent.value, void 0, void 0);
		});
		const changeTab = (index) => {
			selectedResponseIndex.value = index;
			selectedExampleKey.value = resolveExampleKey(__props.selectedExample);
		};
		const showSchema = ref(false);
		return (_ctx, _cache) => {
			return statusCodesWithContent.value.length ? (openBlock(), createBlock(unref(ScalarCard), {
				key: 0,
				"aria-label": unref(translate)("response.exampleResponses"),
				class: "response-card",
				role: "region"
			}, {
				default: withCtx(() => [
					createVNode(ExampleResponseTabList_default, { onChange: changeTab }, {
						actions: withCtx(() => [currentResponseContent.value?.example ? (openBlock(), createElementBlock("button", {
							key: 0,
							class: "code-copy",
							type: "button",
							onClick: _cache[0] || (_cache[0] = () => unref(copyToClipboard)(currentResponseContent.value?.example))
						}, [createVNode(unref(ScalarIcon), {
							icon: "Clipboard",
							width: "12px"
						})])) : createCommentVNode("", true), currentResponseContent.value?.schema ? (openBlock(), createElementBlock("label", _hoisted_1, [
							createTextVNode(toDisplayString(unref(translate)("response.showSchema")) + " ", 1),
							withDirectives(createElementVNode("input", {
								"onUpdate:modelValue": _cache[1] || (_cache[1] = ($event) => showSchema.value = $event),
								"aria-controls": unref(id),
								class: "scalar-card-checkbox-input",
								type: "checkbox"
							}, null, 8, _hoisted_2), [[vModelCheckbox, showSchema.value]]),
							_cache[2] || (_cache[2] = createElementVNode("span", { class: "scalar-card-checkbox-checkmark" }, null, -1))
						])) : createCommentVNode("", true)]),
						default: withCtx(() => [(openBlock(true), createElementBlock(Fragment, null, renderList(statusCodesWithContent.value, (statusCode) => {
							return openBlock(), createBlock(ExampleResponseTab_default, {
								key: statusCode,
								"aria-controls": unref(id)
							}, {
								default: withCtx(() => [createVNode(ScreenReader_default, null, {
									default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("response.status")) + ":", 1)]),
									_: 1
								}), createTextVNode(" " + toDisplayString(statusCode), 1)]),
								_: 2
							}, 1032, ["aria-controls"]);
						}), 128))]),
						_: 1
					}),
					createVNode(unref(ScalarCardSection), { class: "grid flex-1" }, {
						default: withCtx(() => [currentResponseContent.value?.schema && showSchema.value ? (openBlock(), createBlock(ExampleSchema_default, {
							key: 0,
							id: unref(id),
							schema: currentResponseContent.value?.schema
						}, null, 8, ["id", "schema"])) : (openBlock(), createBlock(ExampleResponse_default, {
							key: 1,
							id: unref(id),
							example: currentExample.value,
							response: currentResponseContent.value
						}, null, 8, [
							"id",
							"example",
							"response"
						]))]),
						_: 1
					}),
					currentResponse.value?.description || hasMultipleExamples.value ? (openBlock(), createBlock(unref(ScalarCardFooter), {
						key: 0,
						class: "response-card-footer"
					}, {
						default: withCtx(() => [hasMultipleExamples.value ? (openBlock(), createBlock(unref(ExamplePicker), {
							key: 0,
							class: "response-example-selector px-0",
							examples: currentResponseContent.value?.examples,
							modelValue: selectedExampleKey.value,
							"onUpdate:modelValue": selectExample
						}, null, 8, ["examples", "modelValue"])) : createCommentVNode("", true), createElementVNode("div", _hoisted_3, [currentResponse.value?.description ? (openBlock(), createBlock(unref(ScalarMarkdown), {
							key: 0,
							class: "response-description-markdown",
							value: currentResponse.value.description
						}, null, 8, ["value"])) : createCommentVNode("", true)])]),
						_: 1
					})) : createCommentVNode("", true)
				]),
				_: 1
			}, 8, ["aria-label"])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { ExampleResponses_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ExampleResponses.vue.script.js.map