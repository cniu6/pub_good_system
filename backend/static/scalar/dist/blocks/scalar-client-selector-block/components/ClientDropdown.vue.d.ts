import { type ClientOptionGroup } from '@scalar/blocks/code-example';
import { type WorkspaceEventBus } from '@scalar/workspace-store/events';
type __VLS_Props = {
    /** Client options */
    clientOptions: ClientOptionGroup[];
    /** The currently selected Http Client (a built-in client id or a custom sample id) */
    selectedClient?: string;
    /** Event bus */
    eventBus: WorkspaceEventBus;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=ClientDropdown.vue.d.ts.map