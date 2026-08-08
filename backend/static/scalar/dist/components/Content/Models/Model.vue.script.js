import { useIntersection } from "../../../hooks/use-intersection.js";
import ClassicLayout_default from "./components/ClassicLayout.vue.js";
import ModernLayout_default from "./components/ModernLayout.vue.js";
import { createBlock, createCommentVNode, createElementBlock, defineComponent, openBlock, useTemplateRef } from "vue";
//#region src/components/Content/Models/Model.vue?vue&type=script&setup=true&lang.ts
var Model_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Model",
	props: {
		id: {},
		name: {},
		options: {},
		schema: {},
		isCollapsed: { type: Boolean },
		eventBus: {},
		document: {}
	},
	setup(__props) {
		const section = useTemplateRef("section");
		useIntersection(section, () => __props.eventBus?.emit("intersecting:nav-item", { id: __props.id }));
		return (_ctx, _cache) => {
			return __props.schema ? (openBlock(), createElementBlock("div", {
				key: 0,
				ref_key: "section",
				ref: section
			}, [__props.options.layout === "classic" ? (openBlock(), createBlock(ClassicLayout_default, {
				key: 0,
				id: __props.id,
				document: __props.document,
				eventBus: __props.eventBus,
				isCollapsed: __props.isCollapsed,
				name: __props.name,
				options: __props.options,
				schema: __props.schema
			}, null, 8, [
				"id",
				"document",
				"eventBus",
				"isCollapsed",
				"name",
				"options",
				"schema"
			])) : (openBlock(), createBlock(ModernLayout_default, {
				key: 1,
				id: __props.id,
				document: __props.document,
				eventBus: __props.eventBus,
				isCollapsed: __props.isCollapsed,
				name: __props.name,
				options: __props.options,
				schema: __props.schema
			}, null, 8, [
				"id",
				"document",
				"eventBus",
				"isCollapsed",
				"name",
				"options",
				"schema"
			]))], 512)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { Model_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Model.vue.script.js.map