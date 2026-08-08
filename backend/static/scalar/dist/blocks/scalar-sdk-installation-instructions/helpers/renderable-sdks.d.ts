import type { XScalarSdkInstallation } from '@scalar/workspace-store/schemas/extensions/document/x-scalar-sdk-installation';
/** The array shape stored under `x-scalar-sdk-installation`. */
type SdkInstallationList = NonNullable<XScalarSdkInstallation['x-scalar-sdk-installation']>;
/** A renderable SDK entry whose `description` is the resolved Markdown to show. */
type RenderableSdk = {
    lang: string;
    description: string;
};
/**
 * The SDK installation entries that actually have something to render, each
 * resolved to a single Markdown `description`.
 *
 * A `description` is the promoted content, but the legacy `source` install
 * command is still supported: when both are present it is appended to the
 * description as a fenced code block, and when only `source` is present it
 * becomes the description on its own. Entries that carry neither are ignored so
 * the UI can fall back to the generic client selector instead of showing an
 * empty card. Both the gate in `Content.vue` and the tab list in the block rely
 * on this, so the "has instructions" rule lives in one place.
 *
 * The value comes straight from an untrusted OpenAPI document, so anything
 * malformed — a non-array extension, a non-object entry, or an entry missing a
 * string `lang` (the tab label and icon key) and any content — is treated as
 * "no instructions" rather than allowed to throw at render time.
 */
export declare const getRenderableSdks: (xScalarSdkInstallation: SdkInstallationList | undefined) => RenderableSdk[];
export {};
//# sourceMappingURL=renderable-sdks.d.ts.map