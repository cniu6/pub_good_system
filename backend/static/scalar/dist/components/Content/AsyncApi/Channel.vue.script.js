import Section_default from "../../Section/Section.vue.js";
import SectionContainer_default from "../../Section/SectionContainer.vue.js";
import SectionContainerAccordion_default from "../../Section/SectionContainerAccordion.vue.js";
import SectionContent_default from "../../Section/SectionContent.vue.js";
import SectionHeader_default from "../../Section/SectionHeader.vue.js";
import SectionHeaderTag_default from "../../Section/SectionHeaderTag.vue.js";
import Anchor_default from "../../Anchor/Anchor.vue.js";
import ParameterList_default from "../../../features/Operation/components/ParameterList.vue.js";
import AsyncApiLabels_default from "./AsyncApiLabels.vue.js";
import { adaptAsyncApiParameters } from "./helpers/adapt-async-api-parameters.js";
import { resolveSchemaRenderOptions } from "./helpers/async-api-render-options.js";
import { filterChildrenByType } from "./helpers/filter-children-by-type.js";
import { getChannelServerLabels } from "./helpers/get-async-api-labels.js";
import { pickHeading } from "./helpers/pick-heading.js";
import { resolveAsyncApiChannel } from "./helpers/resolve-async-api-nodes.js";
import Operation_default from "./Operation.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createTextVNode, createVNode, defineComponent, openBlock, renderList, toDisplayString, unref, useId, withCtx } from "vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/components/Content/AsyncApi/Channel.vue?vue&type=script&setup=true&lang.ts
var Channel_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Channel",
	props: {
		channel: {},
		document: {},
		layout: {},
		isCollapsed: { type: Boolean },
		eventBus: {},
		options: {},
		expandedItems: { default: () => ({}) },
		level: { default: 0 }
	},
	setup(__props) {
		const headerId = useId();
		/**
		* Resolve the channel from the document so we can read its description.
		* The navigation entry only carries the address and identifying keys.
		*/
		const resolvedChannel = computed(() => resolveAsyncApiChannel(__props.document, __props.channel.channelName));
		const description = computed(() => resolvedChannel.value?.description ?? "");
		/**
		* Heading shown above each channel section.
		*
		* Prefers the human-friendly `channel.title` when it is set, then falls back
		* to `channel.address`. `channel.channelAddress` already encodes the
		* address-or-key fallback during navigation traversal, so we only need to
		* overlay `title` on top of it.
		*/
		const headingText = computed(() => pickHeading(resolvedChannel.value?.title, __props.channel.channelAddress));
		/**
		* Channel address parameters mapped into the OpenAPI parameter shape so we can reuse the shared
		* `ParameterList` component instead of building a dedicated AsyncAPI renderer.
		*/
		const parameters = computed(() => adaptAsyncApiParameters(resolvedChannel.value?.parameters));
		/** Fill in defaults so the shared renderer always receives a complete options object. */
		const parameterListOptions = computed(() => ({
			hideModels: __props.options?.hideModels ?? false,
			...resolveSchemaRenderOptions(__props.options)
		}));
		/**
		* Server-name and protocol labels for the channel, resolved from `document.servers`
		* (restricted to `channel.servers` when declared) and rendered as pills in the header.
		*/
		const labels = computed(() => getChannelServerLabels(__props.document, resolvedChannel.value));
		/** Operations that target this channel, rendered nested beneath the channel content. */
		const operations = computed(() => filterChildrenByType(__props.channel.children, "asyncapi-operation"));
		return (_ctx, _cache) => {
			return __props.layout === "classic" ? (openBlock(), createBlock(unref(SectionContainerAccordion_default), {
				key: 0,
				"aria-label": headingText.value,
				class: "channel-section",
				modelValue: !__props.isCollapsed,
				"onUpdate:modelValue": _cache[1] || (_cache[1] = (value) => __props.eventBus?.emit("toggle:nav-item", {
					id: __props.channel.id,
					open: value
				}))
			}, {
				title: withCtx(() => [
					createVNode(unref(SectionHeader_default), { class: "channel-name" }, {
						default: withCtx(() => [createVNode(unref(Anchor_default), { onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.channel.id })) }, {
							default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), { level: 2 }, {
								default: withCtx(() => [createTextVNode(toDisplayString(headingText.value), 1)]),
								_: 1
							})]),
							_: 1
						})]),
						_: 1
					}),
					createVNode(AsyncApiLabels_default, {
						class: "channel-labels",
						protocols: labels.value.protocols,
						servers: labels.value.servers
					}, null, 8, ["protocols", "servers"]),
					createVNode(unref(ScalarMarkdown), {
						class: "channel-description",
						value: description.value,
						withImages: ""
					}, null, 8, ["value"])
				]),
				default: withCtx(() => [parameters.value.length ? (openBlock(), createBlock(ParameterList_default, {
					key: 0,
					eventBus: __props.eventBus,
					options: parameterListOptions.value,
					parameters: parameters.value
				}, {
					title: withCtx(() => [..._cache[4] || (_cache[4] = [createTextVNode("Parameters", -1)])]),
					_: 1
				}, 8, [
					"eventBus",
					"options",
					"parameters"
				])) : createCommentVNode("", true), (openBlock(true), createElementBlock(Fragment, null, renderList(operations.value, (operation) => {
					return openBlock(), createBlock(Operation_default, {
						key: operation.id,
						document: __props.document,
						eventBus: __props.eventBus,
						expandedItems: __props.expandedItems,
						operation,
						options: __props.options
					}, null, 8, [
						"document",
						"eventBus",
						"expandedItems",
						"operation",
						"options"
					]);
				}), 128))]),
				_: 1
			}, 8, ["aria-label", "modelValue"])) : (openBlock(), createBlock(unref(SectionContainer_default), {
				key: 1,
				"aria-labelledby": unref(headerId),
				omit: __props.level !== 0,
				role: "region"
			}, {
				default: withCtx(() => [createVNode(unref(Section_default), {
					id: __props.channel.id,
					role: "none",
					onIntersecting: _cache[3] || (_cache[3] = () => __props.eventBus?.emit("intersecting:nav-item", { id: __props.channel.id }))
				}, {
					default: withCtx(() => [createVNode(unref(SectionHeader_default), null, {
						default: withCtx(() => [createVNode(unref(Anchor_default), { onCopyAnchorUrl: _cache[2] || (_cache[2] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.channel.id })) }, {
							default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), {
								id: unref(headerId),
								level: 2
							}, {
								default: withCtx(() => [createTextVNode(toDisplayString(headingText.value), 1)]),
								_: 1
							}, 8, ["id"])]),
							_: 1
						})]),
						_: 1
					}), createVNode(unref(SectionContent_default), null, {
						default: withCtx(() => [
							createVNode(AsyncApiLabels_default, {
								class: "channel-labels",
								protocols: labels.value.protocols,
								servers: labels.value.servers
							}, null, 8, ["protocols", "servers"]),
							createVNode(unref(ScalarMarkdown), {
								value: description.value,
								withImages: ""
							}, null, 8, ["value"]),
							parameters.value.length ? (openBlock(), createBlock(ParameterList_default, {
								key: 0,
								eventBus: __props.eventBus,
								options: parameterListOptions.value,
								parameters: parameters.value
							}, {
								title: withCtx(() => [..._cache[5] || (_cache[5] = [createTextVNode("Parameters", -1)])]),
								_: 1
							}, 8, [
								"eventBus",
								"options",
								"parameters"
							])) : createCommentVNode("", true),
							(openBlock(true), createElementBlock(Fragment, null, renderList(operations.value, (operation) => {
								return openBlock(), createBlock(Operation_default, {
									key: operation.id,
									document: __props.document,
									eventBus: __props.eventBus,
									expandedItems: __props.expandedItems,
									operation,
									options: __props.options
								}, null, 8, [
									"document",
									"eventBus",
									"expandedItems",
									"operation",
									"options"
								]);
							}), 128))
						]),
						_: 1
					})]),
					_: 1
				}, 8, ["id"])]),
				_: 1
			}, 8, ["aria-labelledby", "omit"]));
		};
	}
});
//#endregion
export { Channel_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Channel.vue.script.js.map