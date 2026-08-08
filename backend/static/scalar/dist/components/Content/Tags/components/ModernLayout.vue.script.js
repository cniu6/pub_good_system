import { useLocalization } from "../../../../features/localization/use-localization.js";
import SectionContainer_default from "../../../Section/SectionContainer.vue.js";
import ShowMoreButton_default from "../../../ShowMoreButton.vue.js";
import TagSection_default from "./TagSection.vue.js";
import { computed, createBlock, createCommentVNode, createElementBlock, defineComponent, openBlock, renderSlot, unref, useId, withCtx } from "vue";
//#region src/components/Content/Tags/components/ModernLayout.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 2,
	class: "contents divide-y"
};
var ModernLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ModernLayout",
	props: {
		tag: {},
		moreThanOneTag: { type: Boolean },
		isCollapsed: { type: Boolean },
		eventBus: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		const headerId = useId();
		const moreThanOneDefaultTag = computed(() => __props.moreThanOneTag || __props.tag?.title !== "default" || __props.tag?.description !== "");
		const hasChildren = computed(() => (__props.tag?.children?.length ?? 0) > 0);
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(SectionContainer_default), {
				"aria-labelledby": unref(headerId),
				class: "tag-section-container",
				role: "region"
			}, {
				default: withCtx(() => [
					moreThanOneDefaultTag.value ? (openBlock(), createBlock(TagSection_default, {
						key: 0,
						eventBus: __props.eventBus,
						headerId: unref(headerId),
						isCollapsed: __props.isCollapsed,
						tag: __props.tag
					}, null, 8, [
						"eventBus",
						"headerId",
						"isCollapsed",
						"tag"
					])) : createCommentVNode("", true),
					__props.isCollapsed && __props.moreThanOneTag && hasChildren.value ? (openBlock(), createBlock(ShowMoreButton_default, {
						key: 1,
						id: __props.tag.id,
						"aria-label": unref(translate)("navigation.showAllEndpoints", { name: __props.tag.title }),
						onClick: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("toggle:nav-item", {
							id: __props.tag.id,
							open: true
						}))
					}, null, 8, ["id", "aria-label"])) : createCommentVNode("", true),
					!(__props.isCollapsed && __props.moreThanOneTag) ? (openBlock(), createElementBlock("div", _hoisted_1, [renderSlot(_ctx.$slots, "default", {}, void 0, true)])) : createCommentVNode("", true)
				]),
				_: 3
			}, 8, ["aria-labelledby"]);
		};
	}
});
//#endregion
export { ModernLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ModernLayout.vue.script.js.map