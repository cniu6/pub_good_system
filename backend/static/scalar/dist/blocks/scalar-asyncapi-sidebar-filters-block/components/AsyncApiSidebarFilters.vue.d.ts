import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import { type Component } from 'vue';
/**
 * AsyncApiSidebarFilters
 *
 * The "Filters" sidebar section shown for AsyncAPI documents. Bundling the
 * native sidebar title and picker pair here keeps the two sidebar layouts
 * (modern and classic) from each repeating the option-building logic.
 *
 * Each picker hides itself when there is nothing to choose from, and the whole
 * section hides when neither picker has a real choice — so passing an OpenAPI
 * document (or `null`) renders nothing.
 */
type __VLS_Props = {
    /** The active document, or `null` for OpenAPI documents (renders nothing). */
    document: AsyncApiDocument | null;
    /** Render element for the sidebar section wrapper. */
    is?: Component | string;
};
type __VLS_ModelProps = {
    /** Selected protocol id; empty string clears the filter. */
    'protocol'?: string;
    /** Selected server name; empty string clears the filter. */
    'server'?: string;
};
type __VLS_PublicProps = __VLS_Props & __VLS_ModelProps;
declare const __VLS_export: import("vue").DefineComponent<__VLS_PublicProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {
    "update:protocol": (value: string) => any;
    "update:server": (value: string) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_PublicProps> & Readonly<{
    "onUpdate:protocol"?: ((value: string) => any) | undefined;
    "onUpdate:server"?: ((value: string) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=AsyncApiSidebarFilters.vue.d.ts.map