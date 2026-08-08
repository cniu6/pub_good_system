import DownloadLink_default from "./DownloadLink.vue.js";
import IntroductionCard_default from "./IntroductionCard.vue.js";
import IntroductionLayout_default from "./IntroductionLayout.vue.js";
import { computed, createBlock, createVNode, defineComponent, openBlock, renderSlot, withCtx } from "vue";
//#region src/blocks/scalar-info-block/components/InfoBlock.vue?vue&type=script&setup=true&lang.ts
var InfoBlock_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "InfoBlock",
	props: {
		id: {},
		specificationVersion: {},
		info: {},
		externalDocs: {},
		documentExtensions: {},
		infoExtensions: {},
		eventBus: {},
		headingSlugGenerator: { type: Function },
		layout: {},
		documentDownloadType: { default: "both" },
		documentUrl: {},
		documentType: {}
	},
	setup(__props) {
		/**
		* Put the selectors in
		* - the after slot for classic layout,
		* - and the aside slot for other layouts.
		*/
		const introCardsSlot = computed(() => __props.layout === "classic" ? "after" : "aside");
		return (_ctx, _cache) => {
			return openBlock(), createBlock(IntroductionLayout_default, {
				id: __props.id,
				documentExtensions: __props.documentExtensions,
				documentType: __props.documentType,
				eventBus: __props.eventBus,
				externalDocs: __props.externalDocs,
				headingSlugGenerator: __props.headingSlugGenerator,
				info: __props.info,
				infoExtensions: __props.infoExtensions,
				specificationVersion: __props.specificationVersion
			}, {
				[introCardsSlot.value]: withCtx(() => [createVNode(IntroductionCard_default, { row: __props.layout === "classic" }, {
					default: withCtx(() => [renderSlot(_ctx.$slots, "selectors")]),
					_: 3
				}, 8, ["row"])]),
				"download-link": withCtx(() => [createVNode(DownloadLink_default, {
					documentDownloadType: __props.documentDownloadType,
					documentType: __props.documentType,
					documentUrl: __props.documentUrl,
					eventBus: __props.eventBus
				}, null, 8, [
					"documentDownloadType",
					"documentType",
					"documentUrl",
					"eventBus"
				])]),
				_: 2
			}, 1032, [
				"id",
				"documentExtensions",
				"documentType",
				"eventBus",
				"externalDocs",
				"headingSlugGenerator",
				"info",
				"infoExtensions",
				"specificationVersion"
			]);
		};
	}
});
//#endregion
export { InfoBlock_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=InfoBlock.vue.script.js.map