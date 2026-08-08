import { useLocalization } from "../../../features/localization/use-localization.js";
import ScreenReader_default from "../../ScreenReader.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, renderList, toDisplayString, unref, withCtx, withModifiers } from "vue";
import { ScalarCard, ScalarCardHeader, ScalarCardSection } from "@scalar/components/card";
//#region src/components/Content/AsyncApi/ChannelsList.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["aria-label"];
var _hoisted_2 = ["onClick"];
var ChannelsList_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ChannelsList",
	props: {
		tag: {},
		eventBus: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* Channels grouped under this tag. The OpenAPI `OperationsList` only knows about
		* `operation`/`webhook` children, so AsyncAPI tags need their own list that links
		* straight to each channel section.
		*/
		const channels = computed(() => __props.tag.children?.filter((child) => child.type === "asyncapi-channel") ?? []);
		return (_ctx, _cache) => {
			return channels.value.length ? (openBlock(), createBlock(unref(ScalarCard), {
				key: 0,
				class: "channels-card"
			}, {
				default: withCtx(() => [createVNode(unref(ScalarCardHeader), { muted: "" }, {
					default: withCtx(() => [createVNode(ScreenReader_default, null, {
						default: withCtx(() => [createTextVNode(toDisplayString(__props.tag.title), 1)]),
						_: 1
					}), createTextVNode(" " + toDisplayString(unref(translate)("navigation.channels")), 1)]),
					_: 1
				}), createVNode(unref(ScalarCardSection), { class: "custom-scroll max-h-[60vh]" }, {
					default: withCtx(() => [createElementVNode("ul", {
						"aria-label": unref(translate)("navigation.channels"),
						class: "channels"
					}, [(openBlock(true), createElementBlock(Fragment, null, renderList(channels.value, (channel) => {
						return openBlock(), createElementBlock("li", {
							key: channel.id,
							class: "contents"
						}, [createElementVNode("a", {
							class: "channel",
							onClick: withModifiers(() => __props.eventBus?.emit("scroll-to:nav-item", { id: channel.id }), ["prevent"])
						}, toDisplayString(channel.title || channel.channelAddress), 9, _hoisted_2)]);
					}), 128))], 8, _hoisted_1)]),
					_: 1
				})]),
				_: 1
			})) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { ChannelsList_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ChannelsList.vue.script.js.map