import { useLocalization } from "../../../../features/localization/use-localization.js";
import Section_default from "../../../Section/Section.vue.js";
import SectionColumn_default from "../../../Section/SectionColumn.vue.js";
import SectionColumns_default from "../../../Section/SectionColumns.vue.js";
import SectionContent_default from "../../../Section/SectionContent.vue.js";
import SectionHeader_default from "../../../Section/SectionHeader.vue.js";
import SectionHeaderTag_default from "../../../Section/SectionHeaderTag.vue.js";
import ScreenReader_default from "../../../ScreenReader.vue.js";
import Anchor_default from "../../../Anchor/Anchor.vue.js";
import SpecificationExtension_default from "../../../../features/specification-extension/SpecificationExtension.vue.js";
import ChannelsList_default from "../../AsyncApi/ChannelsList.vue.js";
import OperationsList_default from "../../../OperationsList/OperationsList.vue.js";
import { computed, createBlock, createCommentVNode, createTextVNode, createVNode, defineComponent, openBlock, toDisplayString, unref, withCtx } from "vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/components/Content/Tags/components/TagSection.vue?vue&type=script&setup=true&lang.ts
var TagSection_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "TagSection",
	props: {
		tag: {},
		headerId: {},
		isCollapsed: { type: Boolean },
		eventBus: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* AsyncAPI tags carry `asyncapi-channel` children instead of `operation`/`webhook`,
		* so they get a dedicated channel list rather than the (empty) operations card.
		*/
		const hasChannels = computed(() => __props.tag.children?.some((child) => child.type === "asyncapi-channel") ?? false);
		return (_ctx, _cache) => {
			return __props.tag ? (openBlock(), createBlock(unref(Section_default), {
				key: 0,
				id: __props.tag.id,
				role: "none",
				onIntersecting: _cache[1] || (_cache[1] = () => __props.eventBus?.emit("intersecting:nav-item", { id: __props.tag.id }))
			}, {
				default: withCtx(() => [
					createVNode(unref(SectionHeader_default), null, {
						default: withCtx(() => [createVNode(unref(Anchor_default), { onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.tag.id })) }, {
							default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), {
								id: __props.headerId,
								level: 2
							}, {
								default: withCtx(() => [createTextVNode(toDisplayString(__props.tag.title) + " ", 1), __props.isCollapsed ? (openBlock(), createBlock(ScreenReader_default, { key: 0 }, {
									default: withCtx(() => [createTextVNode(" (" + toDisplayString(unref(translate)("navigation.collapsed")) + ") ", 1)]),
									_: 1
								})) : createCommentVNode("", true)]),
								_: 1
							}, 8, ["id"])]),
							_: 1
						})]),
						_: 1
					}),
					createVNode(unref(SectionContent_default), null, {
						default: withCtx(() => [createVNode(unref(SectionColumns_default), null, {
							default: withCtx(() => [createVNode(unref(SectionColumn_default), null, {
								default: withCtx(() => [createVNode(unref(ScalarMarkdown), {
									clamp: __props.isCollapsed ? 7 : void 0,
									value: __props.tag?.description ?? "",
									withImages: ""
								}, null, 8, ["clamp", "value"])]),
								_: 1
							}), createVNode(unref(SectionColumn_default), null, {
								default: withCtx(() => [hasChannels.value ? (openBlock(), createBlock(ChannelsList_default, {
									key: 0,
									eventBus: __props.eventBus,
									tag: __props.tag
								}, null, 8, ["eventBus", "tag"])) : (openBlock(), createBlock(unref(OperationsList_default), {
									key: 1,
									eventBus: __props.eventBus,
									tag: __props.tag
								}, null, 8, ["eventBus", "tag"]))]),
								_: 1
							})]),
							_: 1
						})]),
						_: 1
					}),
					createVNode(unref(SpecificationExtension_default), { value: __props.tag.xKeys }, null, 8, ["value"])
				]),
				_: 1
			}, 8, ["id"])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { TagSection_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=TagSection.vue.script.js.map