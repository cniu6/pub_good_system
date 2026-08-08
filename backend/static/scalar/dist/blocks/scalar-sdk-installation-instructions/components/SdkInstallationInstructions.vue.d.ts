import { type WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { XScalarSdkInstallation } from '@scalar/workspace-store/schemas/extensions/document/x-scalar-sdk-installation';
type __VLS_Props = {
    /** Custom SDK installation instructions from `x-scalar-sdk-installation` */
    xScalarSdkInstallation?: XScalarSdkInstallation['x-scalar-sdk-installation'];
    /**
     * The globally selected client id. When it matches one of the SDK languages
     * (as a `custom/<lang>` id) the matching tab is shown as active, keeping the
     * tabs in sync with the operation code samples.
     */
    selectedClient?: string;
    /**
     * Event bus used to broadcast the selected client. Picking a language here
     * switches the operation code samples to that language's custom example, the
     * same channel the generic client selector uses.
     */
    eventBus?: WorkspaceEventBus;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=SdkInstallationInstructions.vue.d.ts.map