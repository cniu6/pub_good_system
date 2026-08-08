import { useIntersection } from "../../../hooks/use-intersection.js";
import SectionHeaderTag_default from "../../Section/SectionHeaderTag.vue.js";
import Anchor_default from "../../Anchor/Anchor.vue.js";
import { filterChildrenByType } from "./helpers/filter-children-by-type.js";
import { pickHeading } from "./helpers/pick-heading.js";
import { resolveAsyncApiOperation } from "./helpers/resolve-async-api-nodes.js";
import OperationScopes_default from "../../../features/Operation/components/OperationScopes.vue.js";
import { getAsyncApiRequiredSecurity } from "./helpers/get-async-api-required-security.js";
import Message_default from "./Message.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, renderList, toDisplayString, unref, useId, useTemplateRef, withCtx } from "vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/components/Content/AsyncApi/Operation.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id"];
var _hoisted_2 = { class: "operation-header" };
var Operation_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Operation",
	props: {
		operation: {},
		document: {},
		eventBus: {},
		options: {},
		expandedItems: { default: () => ({}) }
	},
	setup(__props) {
		const headerId = useId();
		const section = useTemplateRef("section");
		useIntersection(section, () => __props.eventBus?.emit("intersecting:nav-item", { id: __props.operation.id }));
		/**
		* Resolve the operation from the document so we can read its summary/description.
		* Operation traits are merged in (matching the channel connection UI) so trait-only
		* fields render as part of the operation.
		*/
		const resolvedOperation = computed(() => resolveAsyncApiOperation(__props.document, __props.operation.operationName));
		/** Heading prefers title, then summary, then the operation map key. */
		const headingText = computed(() => pickHeading(resolvedOperation.value?.title, __props.operation.title, __props.operation.operationName));
		const description = computed(() => resolvedOperation.value?.description || resolvedOperation.value?.summary || "");
		/** Only the message children of this operation. */
		const messages = computed(() => filterChildrenByType(__props.operation.children, "asyncapi-message"));
		/** OAuth scopes required by this operation, rendered below the description. */
		const requiredSecurity = computed(() => getAsyncApiRequiredSecurity(__props.document, resolvedOperation.value));
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				id: __props.operation.id,
				ref_key: "section",
				ref: section,
				class: normalizeClass(["operation", `operation--${__props.operation.action}`])
			}, [
				createElementVNode("div", _hoisted_2, [createElementVNode("span", { class: normalizeClass(["operation-action", `operation-action--${__props.operation.action}`]) }, toDisplayString(__props.operation.action), 3), createVNode(unref(Anchor_default), { onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.operation.id })) }, {
					default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), {
						id: unref(headerId),
						class: "operation-title",
						level: 3
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(headingText.value), 1)]),
						_: 1
					}, 8, ["id"])]),
					_: 1
				})]),
				description.value ? (openBlock(), createBlock(unref(ScalarMarkdown), {
					key: 0,
					class: "operation-description",
					value: description.value,
					withImages: ""
				}, null, 8, ["value"])) : createCommentVNode("", true),
				createVNode(OperationScopes_default, { requiredSecurity: requiredSecurity.value }, null, 8, ["requiredSecurity"]),
				(openBlock(true), createElementBlock(Fragment, null, renderList(messages.value, (message) => {
					return openBlock(), createBlock(Message_default, {
						key: message.id,
						document: __props.document,
						eventBus: __props.eventBus,
						expandedItems: __props.expandedItems,
						message,
						options: __props.options
					}, null, 8, [
						"document",
						"eventBus",
						"expandedItems",
						"message",
						"options"
					]);
				}), 128))
			], 10, _hoisted_1);
		};
	}
});
//#endregion
export { Operation_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Operation.vue.script.js.map