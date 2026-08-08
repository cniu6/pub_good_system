import { useLocalization } from "../../../features/localization/use-localization.js";
import Section_default from "../../../components/Section/Section.vue.js";
import SectionColumn_default from "../../../components/Section/SectionColumn.vue.js";
import SectionColumns_default from "../../../components/Section/SectionColumns.vue.js";
import SectionContainer_default from "../../../components/Section/SectionContainer.vue.js";
import SectionContent_default from "../../../components/Section/SectionContent.vue.js";
import SectionHeader_default from "../../../components/Section/SectionHeader.vue.js";
import SectionHeaderTag_default from "../../../components/Section/SectionHeaderTag.vue.js";
import SpecificationExtension_default from "../../../features/specification-extension/SpecificationExtension.vue.js";
import InfoDescription_default from "./InfoDescription.vue.js";
import InfoLinks_default from "./InfoLinks.vue.js";
import InfoVersion_default from "./InfoVersion.vue.js";
import IntroductionLoading_default from "./IntroductionLoading.vue.js";
import SpecificationVersion_default from "./SpecificationVersion.vue.js";
import { Fragment, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, renderSlot, toDisplayString, unref, withCtx } from "vue";
//#region src/blocks/scalar-info-block/components/IntroductionLayout.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "flex gap-1.5" };
var _hoisted_2 = { class: "sticky-cards" };
var IntroductionLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "IntroductionLayout",
	props: {
		id: {},
		documentType: {},
		specificationVersion: {},
		info: {},
		externalDocs: {},
		documentExtensions: {},
		infoExtensions: {},
		headingSlugGenerator: { type: Function },
		eventBus: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(SectionContainer_default), null, {
				default: withCtx(() => [createVNode(unref(Section_default), {
					id: __props.id,
					"aria-label": unref(translate)("navigation.introduction"),
					class: "introduction-section z-1 gap-12",
					onIntersecting: _cache[0] || (_cache[0] = () => __props.id && __props.eventBus?.emit("intersecting:nav-item", { id: __props.id }))
				}, {
					default: withCtx(() => [createVNode(unref(SectionContent_default), null, {
						default: withCtx(() => [!__props.info ? (openBlock(), createBlock(IntroductionLoading_default, {
							key: 0,
							hasAside: Boolean(_ctx.$slots.aside)
						}, null, 8, ["hasAside"])) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [
							createElementVNode("div", _hoisted_1, [createVNode(InfoVersion_default, { version: __props.info?.version }, null, 8, ["version"]), createVNode(SpecificationVersion_default, {
								documentType: __props.documentType,
								version: __props.specificationVersion
							}, null, 8, ["documentType", "version"])]),
							createVNode(unref(SectionHeader_default), { tight: "" }, {
								links: withCtx(() => [createVNode(InfoLinks_default, {
									externalDocs: __props.externalDocs,
									info: __props.info
								}, null, 8, ["externalDocs", "info"])]),
								default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), { level: 1 }, {
									default: withCtx(() => [createTextVNode(toDisplayString(__props.info?.title), 1)]),
									_: 1
								})]),
								_: 1
							}),
							createVNode(unref(SectionColumns_default), null, {
								default: withCtx(() => [createVNode(unref(SectionColumn_default), null, {
									default: withCtx(() => [renderSlot(_ctx.$slots, "download-link", {}, void 0, true), createVNode(InfoDescription_default, {
										description: __props.info?.description,
										eventBus: __props.eventBus,
										headingSlugGenerator: __props.headingSlugGenerator
									}, null, 8, [
										"description",
										"eventBus",
										"headingSlugGenerator"
									])]),
									_: 3
								}), _ctx.$slots.aside ? (openBlock(), createBlock(unref(SectionColumn_default), { key: 0 }, {
									default: withCtx(() => [createElementVNode("div", _hoisted_2, [renderSlot(_ctx.$slots, "aside", {}, void 0, true)])]),
									_: 3
								})) : createCommentVNode("", true)]),
								_: 3
							}),
							createVNode(unref(SpecificationExtension_default), { value: __props.documentExtensions }, null, 8, ["value"]),
							createVNode(unref(SpecificationExtension_default), { value: __props.infoExtensions }, null, 8, ["value"])
						], 64))]),
						_: 3
					}), renderSlot(_ctx.$slots, "after", {}, void 0, true)]),
					_: 3
				}, 8, ["id", "aria-label"])]),
				_: 3
			});
		};
	}
});
//#endregion
export { IntroductionLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=IntroductionLayout.vue.script.js.map