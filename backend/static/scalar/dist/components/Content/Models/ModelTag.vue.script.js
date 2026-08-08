import Section_default from "../../Section/Section.vue.js";
import SectionContainer_default from "../../Section/SectionContainer.vue.js";
import SectionContainerAccordion_default from "../../Section/SectionContainerAccordion.vue.js";
import SectionHeader_default from "../../Section/SectionHeader.vue.js";
import SectionHeaderTag_default from "../../Section/SectionHeaderTag.vue.js";
import ShowMoreButton_default from "../../ShowMoreButton.vue.js";
import { createBlock, createTextVNode, createVNode, defineComponent, openBlock, renderSlot, toDisplayString, unref, withCtx } from "vue";
import { DEFAULT_MODELS_SECTION_LABEL } from "@scalar/types/api-reference";
//#region src/components/Content/Models/ModelTag.vue?vue&type=script&setup=true&lang.ts
var ModelTag_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ModelTag",
	props: {
		id: {},
		isCollapsed: { type: Boolean },
		eventBus: {},
		layout: {},
		modelsSectionLabel: { default: () => DEFAULT_MODELS_SECTION_LABEL }
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return __props.layout === "modern" ? (openBlock(), createBlock(SectionContainer_default, { key: 0 }, {
				default: withCtx(() => [createVNode(unref(Section_default), {
					id: __props.id,
					"aria-label": __props.modelsSectionLabel,
					onIntersecting: _cache[1] || (_cache[1] = () => __props.eventBus?.emit("intersecting:nav-item", { id: __props.id }))
				}, {
					default: withCtx(() => [createVNode(unref(SectionHeader_default), null, {
						default: withCtx(() => [createVNode(SectionHeaderTag_default, { level: 2 }, {
							default: withCtx(() => [createTextVNode(toDisplayString(__props.modelsSectionLabel), 1)]),
							_: 1
						})]),
						_: 1
					}), !__props.isCollapsed ? renderSlot(_ctx.$slots, "default", {}, void 0, void 0, 0) : (openBlock(), createBlock(ShowMoreButton_default, {
						key: 1,
						id: __props.id,
						class: "top-0",
						onClick: _cache[0] || (_cache[0] = () => __props.eventBus.emit("toggle:nav-item", {
							id: __props.id,
							open: true
						}))
					}, null, 8, ["id"]))]),
					_: 3
				}, 8, ["id", "aria-label"])]),
				_: 3
			})) : (openBlock(), createBlock(SectionContainerAccordion_default, {
				key: 1,
				"aria-label": __props.modelsSectionLabel,
				class: "pb-12",
				modelValue: !__props.isCollapsed,
				"onUpdate:modelValue": _cache[2] || (_cache[2] = () => __props.eventBus?.emit("toggle:nav-item", {
					id: __props.id,
					open: __props.isCollapsed
				}))
			}, {
				title: withCtx(() => [createVNode(unref(SectionHeader_default), { level: 2 }, {
					default: withCtx(() => [createTextVNode(toDisplayString(__props.modelsSectionLabel), 1)]),
					_: 1
				})]),
				default: withCtx(() => [renderSlot(_ctx.$slots, "default")]),
				_: 3
			}, 8, ["aria-label", "modelValue"]));
		};
	}
});
//#endregion
export { ModelTag_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ModelTag.vue.script.js.map