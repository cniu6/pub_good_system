import ClassicLayout_default from "./components/ClassicLayout.vue.js";
import ModernLayout_default from "./components/ModernLayout.vue.js";
import { createBlock, defineComponent, openBlock, renderSlot, withCtx } from "vue";
//#region src/components/Content/Tags/Tag.vue?vue&type=script&setup=true&lang.ts
var Tag_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Tag",
	props: {
		tag: {},
		layout: {},
		moreThanOneTag: { type: Boolean },
		isCollapsed: { type: Boolean },
		eventBus: {}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return __props.layout === "classic" ? (openBlock(), createBlock(ClassicLayout_default, {
				key: 0,
				eventBus: __props.eventBus,
				isCollapsed: __props.isCollapsed,
				layout: __props.layout,
				tag: __props.tag
			}, {
				default: withCtx(() => [renderSlot(_ctx.$slots, "default")]),
				_: 3
			}, 8, [
				"eventBus",
				"isCollapsed",
				"layout",
				"tag"
			])) : (openBlock(), createBlock(ModernLayout_default, {
				key: 1,
				eventBus: __props.eventBus,
				isCollapsed: __props.isCollapsed,
				layout: __props.layout,
				moreThanOneTag: __props.moreThanOneTag,
				tag: __props.tag
			}, {
				default: withCtx(() => [renderSlot(_ctx.$slots, "default")]),
				_: 3
			}, 8, [
				"eventBus",
				"isCollapsed",
				"layout",
				"moreThanOneTag",
				"tag"
			]));
		};
	}
});
//#endregion
export { Tag_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Tag.vue.script.js.map