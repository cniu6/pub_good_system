import { useIntersection } from "../../../hooks/use-intersection.js";
import SectionAccordion_default from "../../Section/SectionAccordion.vue.js";
import SectionHeaderTag_default from "../../Section/SectionHeaderTag.vue.js";
import Anchor_default from "../../Anchor/Anchor.vue.js";
import Schema_default from "../Schema/Schema.vue.js";
import { getAsyncApiMessageHeadersSchema, getAsyncApiMessagePayloadSchema } from "../../../helpers/get-async-api-message-payload-schema.js";
import AsyncApiLabels_default from "./AsyncApiLabels.vue.js";
import { resolveSchemaRenderOptions } from "./helpers/async-api-render-options.js";
import { getChannelServerLabels } from "./helpers/get-async-api-labels.js";
import { pickHeading } from "./helpers/pick-heading.js";
import { resolveAsyncApiChannel, resolveAsyncApiMessage } from "./helpers/resolve-async-api-nodes.js";
import { computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, ref, toDisplayString, unref, useId, useTemplateRef, watch, withCtx } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/components/Content/AsyncApi/Message.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id"];
var _hoisted_2 = { class: "message-heading" };
var _hoisted_3 = {
	key: 1,
	class: "message-schema"
};
var _hoisted_4 = {
	key: 2,
	class: "message-schema"
};
var Message_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Message",
	props: {
		message: {},
		document: {},
		eventBus: {},
		options: {},
		expandedItems: { default: () => ({}) }
	},
	setup(__props) {
		const headerId = useId();
		const section = useTemplateRef("section");
		useIntersection(section, () => __props.eventBus?.emit("intersecting:nav-item", { id: __props.message.id }));
		/**
		* Resolve the message from the channel it lives on. The navigation entry only
		* carries the identifying keys, so we walk `document.channels[channelName].messages`.
		*/
		const resolvedMessage = computed(() => resolveAsyncApiMessage(__props.document, __props.message.channelName, __props.message.messageName));
		/** Heading prefers the human-friendly title, falling back to the message map key. */
		const headingText = computed(() => pickHeading(resolvedMessage.value?.title, __props.message.title, __props.message.messageName));
		const description = computed(() => resolvedMessage.value?.description || resolvedMessage.value?.summary || "");
		/**
		* Protocol keys declared directly on the message's protocol-specific `bindings`
		* (for example `ws`, `kafka`). The bindings object may be a `$ref`, so it is resolved
		* before reading the keys. Keys are lowercased to match the server protocols (which the
		* helper normalizes), so the union below de-duplicates case-insensitively.
		*/
		const messageBindingProtocols = computed(() => {
			const bindings = resolvedMessage.value?.bindings;
			if (!bindings) return [];
			const resolved = getResolvedRef(bindings);
			return Object.entries(resolved).filter(([, value]) => value != null).map(([protocol]) => protocol.toLowerCase());
		});
		/**
		* Protocol labels for the message. A message is carried over whatever protocols its
		* channel's servers speak, so we start from the channel's server protocols and union in
		* any extra protocols the message declares its own bindings for. Without this a
		* multi-protocol message would only surface the single protocol it happens to declare a
		* binding for (or none at all).
		*/
		const protocolLabels = computed(() => {
			const channel = resolveAsyncApiChannel(__props.document, __props.message.channelName);
			const { protocols } = getChannelServerLabels(__props.document, channel);
			return [.../* @__PURE__ */ new Set([...protocols, ...messageBindingProtocols.value])];
		});
		/** Payload schema, unwrapped from `$ref`s and Multi Format Schema wrappers. */
		const payloadSchema = computed(() => resolvedMessage.value ? getAsyncApiMessagePayloadSchema(resolvedMessage.value) : void 0);
		/** Header schema, unwrapped the same way as the payload. */
		const headersSchema = computed(() => resolvedMessage.value ? getAsyncApiMessageHeadersSchema(resolvedMessage.value) : void 0);
		/** Fill in defaults so the shared Schema renderer always receives a complete options object. */
		const schemaOptions = computed(() => ({
			hideReadOnly: false,
			...resolveSchemaRenderOptions(__props.options)
		}));
		/**
		* Accordion open state. Kept locally so clicking the header always toggles
		* immediately, and seeded/synced from the shared sidebar expansion map so
		* expanding the message in the sidebar (or deep-linking to it) opens it here too.
		*/
		const isExpanded = ref(__props.expandedItems[__props.message.id] ?? false);
		watch(() => __props.expandedItems[__props.message.id], (value) => {
			if (value !== void 0) isExpanded.value = value;
		});
		/** Toggling the accordion updates local state and keeps the sidebar in sync. */
		const onToggle = (open) => {
			isExpanded.value = open;
			__props.eventBus?.emit("toggle:nav-item", {
				id: __props.message.id,
				open
			});
		};
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				id: __props.message.id,
				ref_key: "section",
				ref: section,
				class: "message"
			}, [createVNode(unref(SectionAccordion_default), {
				class: "message-accordion",
				modelValue: isExpanded.value,
				"onUpdate:modelValue": onToggle
			}, {
				title: withCtx(() => [createVNode(unref(Anchor_default), { onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.message.id })) }, {
					default: withCtx(() => [createElementVNode("span", _hoisted_2, [createVNode(unref(SectionHeaderTag_default), {
						id: unref(headerId),
						class: "message-title",
						level: 4
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(headingText.value), 1)]),
						_: 1
					}, 8, ["id"]), createVNode(AsyncApiLabels_default, { protocols: protocolLabels.value }, null, 8, ["protocols"])])]),
					_: 1
				})]),
				default: withCtx(() => [
					description.value ? (openBlock(), createBlock(unref(ScalarMarkdown), {
						key: 0,
						class: "message-description",
						value: description.value,
						withImages: ""
					}, null, 8, ["value"])) : createCommentVNode("", true),
					headersSchema.value ? (openBlock(), createElementBlock("div", _hoisted_3, [_cache[1] || (_cache[1] = createElementVNode("div", { class: "message-schema-title" }, "Headers", -1)), createVNode(unref(Schema_default), {
						compact: "",
						eventBus: __props.eventBus,
						name: "Headers",
						noncollapsible: "",
						options: schemaOptions.value,
						schema: headersSchema.value
					}, null, 8, [
						"eventBus",
						"options",
						"schema"
					])])) : createCommentVNode("", true),
					payloadSchema.value ? (openBlock(), createElementBlock("div", _hoisted_4, [_cache[2] || (_cache[2] = createElementVNode("div", { class: "message-schema-title" }, "Payload", -1)), createVNode(unref(Schema_default), {
						compact: "",
						eventBus: __props.eventBus,
						name: "Payload",
						noncollapsible: "",
						options: schemaOptions.value,
						schema: payloadSchema.value
					}, null, 8, [
						"eventBus",
						"options",
						"schema"
					])])) : createCommentVNode("", true)
				]),
				_: 1
			}, 8, ["modelValue"])], 8, _hoisted_1);
		};
	}
});
//#endregion
export { Message_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Message.vue.script.js.map