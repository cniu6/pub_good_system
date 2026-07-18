<template>
  <n-card :title="t('adminSettings.title')" :bordered="false">
    <n-spin :show="loading">
      <n-tabs v-model:value="topTab" type="line" animated>
        <n-tab-pane name="system-config" :tab="t('adminSettings.systemConfig')">
          <n-tabs v-model:value="systemSubTab" type="line" placement="left" animated>
            <n-tab-pane name="basic" :tab="t('adminSettings.basicSettings')">
              <n-space vertical>
                <n-form :model="basicForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.siteName')">
                    <n-input v-model:value="basicForm.site_name" :placeholder="t('adminSettings.siteNamePlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.siteDesc')">
                    <n-input v-model:value="basicForm.site_desc" type="textarea" :placeholder="t('adminSettings.siteDescPlaceholder')" :rows="3" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.siteLogo')">
                    <n-input v-model:value="basicForm.site_logo" :placeholder="t('adminSettings.siteLogoPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.copyright')">
                    <n-input v-model:value="basicForm.copyright" :placeholder="t('adminSettings.copyrightPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.icp')">
                    <n-input v-model:value="basicForm.icp" :placeholder="t('adminSettings.icpPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.version')">
                    <n-input v-model:value="basicForm.version" :placeholder="t('adminSettings.versionPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.defaultLang')">
                    <n-select v-model:value="basicForm.default_lang" :options="langOptions" :placeholder="t('adminSettings.defaultLangPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.frontendUrl')">
                    <n-input v-model:value="basicForm.frontend_url" :placeholder="t('adminSettings.frontendUrlPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.backendApiUrl')">
                    <n-input v-model:value="basicForm.backend_api_url" :placeholder="t('adminSettings.backendApiUrlPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.userRegistration')">
                    <n-space align="center">
                      <n-switch
                        :value="basicForm.allow_register"
                        :loading="switchLoading.allow_register"
                        @update:value="handleUpdateAllowRegister"
                      />
                      <n-text depth="3">{{ basicForm.allow_register ? t('adminSettings.allowRegister') : t('adminSettings.disallowRegister') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingBasic" @click="handleSaveBasic">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="email" :tab="t('adminSettings.emailSettings')">
              <n-space vertical>
                <n-form :model="emailForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.emailVerification')">
                    <n-space align="center">
                      <n-switch
                        :value="emailForm.email_verify_enabled"
                        :loading="switchLoading.email_verify_enabled"
                        @update:value="handleUpdateEmailVerifyEnabled"
                      />
                      <n-text depth="3">{{ emailForm.email_verify_enabled ? t('adminSettings.emailVerifyEnabled') : t('adminSettings.emailVerifyDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.smtpHost')">
                    <n-input v-model:value="emailForm.smtp_host" :placeholder="t('adminSettings.smtpHostPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpPort')">
                    <n-input-number v-model:value="emailForm.smtp_port" :min="1" :max="65535" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpUsername')">
                    <n-input v-model:value="emailForm.smtp_username" :placeholder="t('adminSettings.smtpUsernamePlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpPassword')">
                    <n-input
                      v-model:value="emailForm.smtp_password"
                      type="password"
                      show-password-on="click"
                      :placeholder="t('adminSettings.smtpPasswordPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.systemEmailName')">
                    <n-input v-model:value="emailForm.system_email_name" :placeholder="t('adminSettings.systemEmailNamePlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpSSL')">
                    <n-space align="center">
                      <n-switch :value="emailForm.smtp_ssl" :loading="switchLoading.smtp_ssl" @update:value="handleUpdateSmtpSSL" />
                      <n-text depth="3">{{ emailForm.smtp_ssl ? t('adminSettings.sslEnabled') : t('adminSettings.sslDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item>
                    <n-space>
                      <n-button type="primary" :loading="savingEmail" @click="handleSaveEmail">{{ t('adminSettings.save') }}</n-button>
                      <n-button :loading="testingEmail" @click="handleTestEmail">{{ t('adminSettings.sendTestEmail') }}</n-button>
                    </n-space>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="sms" :tab="t('adminSettings.smsSettings')">
              <n-space vertical>
                <n-form :model="smsForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.smsVerification')">
                    <n-space align="center">
                      <n-switch
                        :value="smsForm.sms_verify_enabled"
                        :loading="switchLoading.sms_verify_enabled"
                        @update:value="handleUpdateSmsVerifyEnabled"
                      />
                      <n-text depth="3">{{ smsForm.sms_verify_enabled ? t('adminSettings.smsVerifyEnabled') : t('adminSettings.smsVerifyDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.smsProvider')">
                    <n-select
                      v-model:value="smsForm.sms_provider"
                      :options="smsProviderOptions"
                      :placeholder="t('adminSettings.smsProviderPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smsAccessKey')">
                    <n-input
                      v-model:value="smsForm.sms_access_key"
                      :placeholder="smsAccessKeyPlaceholder"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smsSecretKey')">
                    <n-input
                      v-model:value="smsForm.sms_secret_key"
                      type="password"
                      show-password-on="click"
                      :placeholder="smsSecretKeyPlaceholder"
                    />
                  </n-form-item>
                  <n-form-item v-if="smsProviderNeedsSignName" :label="t('adminSettings.smsSignName')">
                    <n-input v-model:value="smsForm.sms_sign_name" :placeholder="t('adminSettings.smsSignNamePlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsProviderNeedsTemplateCode" :label="smsTemplateLabel">
                    <n-input v-model:value="smsForm.sms_template_code" :placeholder="smsTemplatePlaceholder" />
                  </n-form-item>
                  <n-form-item v-if="smsProviderNeedsTemplateCode" :label="smsTemplateEnLabel">
                    <n-input v-model:value="smsForm.sms_template_code_en" :placeholder="smsTemplateEnPlaceholder" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'aliyun'" :label="t('adminSettings.smsRegion')">
                    <n-input v-model:value="smsForm.sms_region" :placeholder="t('adminSettings.smsRegionPlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'tencent'" :label="t('adminSettings.smsSdkAppId')">
                    <n-input v-model:value="smsForm.sms_sdk_app_id" :placeholder="t('adminSettings.smsSdkAppIdPlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'custom'" :label="t('adminSettings.smsEndpoint')">
                    <n-input v-model:value="smsForm.sms_endpoint" :placeholder="t('adminSettings.smsEndpointPlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'custom'" :label="t('adminSettings.smsBodyFormat')">
                    <n-select
                      v-model:value="smsForm.sms_body_format"
                      :options="smsBodyFormatOptions"
                      :placeholder="t('adminSettings.smsBodyFormatPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingSms" @click="handleSaveSms">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" :title="t('adminSettings.tip')" :bordered="false">
                  {{ t('adminSettings.smsAlert') }}
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="security" :tab="t('adminSettings.securitySettings')">
              <n-space vertical>
                <n-form :model="securityForm" label-placement="left" label-width="180px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.geetestEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.geetest_enabled"
                        :loading="switchLoading.geetest_enabled"
                        @update:value="handleUpdateGeetestEnabled"
                      />
                      <n-text depth="3">{{ securityForm.geetest_enabled ? t('adminSettings.enabled') : t('adminSettings.disabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.allowDeleteAccount')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.allow_delete_account"
                        :loading="switchLoading.allow_delete_account"
                        @update:value="handleUpdateAllowDeleteAccount"
                      />
                      <n-text depth="3">{{ securityForm.allow_delete_account ? t('adminSettings.allowDeleteAccountEnabled') : t('adminSettings.allowDeleteAccountDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.geetestCaptchaId')">
                    <n-input v-model:value="securityForm.geetest_captcha_id" :placeholder="t('adminSettings.geetestCaptchaIdPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.geetestCaptchaKey')">
                    <n-input
                      v-model:value="securityForm.geetest_captcha_key"
                      type="password"
                      show-password-on="click"
                      :placeholder="t('adminSettings.geetestCaptchaKeyPlaceholder')"
                    />
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.jwtAccessExpire')">
                    <n-input-number v-model:value="securityForm.jwt_access_expire" :min="300" :step="300" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.jwtRefreshExpire')">
                    <n-input-number v-model:value="securityForm.jwt_refresh_expire" :min="3600" :step="3600" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.loginMaxFailure')">
                    <n-input-number v-model:value="securityForm.login_max_failure" :min="3" :max="20" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.loginLockDuration')">
                    <n-input-number v-model:value="securityForm.login_lock_duration" :min="1" :max="1440" style="width: 100%;" />
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.realnameEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.realname_enabled"
                        :loading="switchLoading.realname_enabled"
                        @update:value="handleUpdateRealnameEnabled"
                      />
                      <n-text depth="3">{{ securityForm.realname_enabled ? t('adminSettings.realnameEnabledText') : t('adminSettings.realnameDisabledText') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameReview')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.realname_review_required"
                        :loading="switchLoading.realname_review_required"
                        @update:value="handleUpdateRealnameReviewRequired"
                      />
                      <n-text depth="3">{{ securityForm.realname_review_required ? t('adminSettings.realnameReviewRequired') : t('adminSettings.realnameReviewNotRequired') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameNotifyText')">
                    <n-input
                      v-model:value="securityForm.realname_notify_text"
                      type="textarea"
                      :placeholder="t('adminSettings.realnameNotifyTextPlaceholder')"
                      :rows="3"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-space>
                      <n-button type="primary" :loading="savingSecurity" @click="handleSaveSecurity">{{ t('adminSettings.saveSettings') }}</n-button>
                      <n-button type="warning" :loading="restartingBackend" @click="handleRestartBackend">{{ t('adminSettings.restartBackend') }}</n-button>
                    </n-space>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="realname-api" :tab="t('adminSettings.realnameApiSettings')">
              <n-space vertical>
                <n-form :model="realnameApiForm" label-placement="left" label-width="180px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.realnameApiEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="realnameApiForm.realname_api_enabled"
                        :loading="switchLoading.realname_api_enabled"
                        @update:value="handleUpdateRealnameApiEnabled"
                      />
                      <n-text depth="3">{{ realnameApiForm.realname_api_enabled ? t('adminSettings.realnameApiEnabledText') : t('adminSettings.realnameApiDisabledText') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.realnameApiProvider')">
                    <n-select
                      v-model:value="realnameApiForm.realname_api_provider"
                      :options="realnameApiProviderOptions"
                      :placeholder="t('adminSettings.realnameApiProviderPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameApiAppKey')">
                    <n-input
                      v-model:value="realnameApiForm.realname_api_app_key"
                      :placeholder="t('adminSettings.realnameApiAppKeyPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameApiAppSecret')">
                    <n-input
                      v-model:value="realnameApiForm.realname_api_app_secret"
                      type="password"
                      show-password-on="click"
                      :placeholder="t('adminSettings.realnameApiAppSecretPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameApiEndpoint')">
                    <n-input
                      v-model:value="realnameApiForm.realname_api_endpoint"
                      :placeholder="t('adminSettings.realnameApiEndpointPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingRealnameApi" @click="handleSaveRealnameApi">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" :title="t('adminSettings.tip')" :bordered="false">
                  <p>{{ t('adminSettings.realnameApiAlert1') }}</p>
                  <p>{{ t('adminSettings.realnameApiAlert2') }}</p>
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="payment" :tab="t('adminSettings.paymentSettings')">
              <n-space vertical>
                <n-form :model="paymentForm" label-placement="left" label-width="140px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.paymentEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="paymentForm.payment_enabled"
                        :loading="switchLoading.payment_enabled"
                        @update:value="handleUpdatePaymentEnabled"
                      />
                      <n-text depth="3">{{ paymentForm.payment_enabled ? t('adminSettings.enabled') : t('adminSettings.disabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.orderExpireMinutes')">
                    <n-input-number v-model:value="paymentForm.payment_order_expire_minutes" :min="1" :max="1440" style="width: 100%;" />
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.withdrawEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="paymentForm.withdraw_enabled"
                        :loading="switchLoading.withdraw_enabled"
                        @update:value="handleUpdateWithdrawEnabled"
                      />
                      <n-text depth="3">{{ paymentForm.withdraw_enabled ? t('adminSettings.withdrawEnabledText') : t('adminSettings.withdrawDisabledText') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.withdrawMinAmount')">
                    <n-input-number v-model:value="paymentForm.withdraw_min_amount" :min="0.01" :precision="2" :step="1" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.withdrawNotifyText')">
                    <n-input
                      v-model:value="paymentForm.withdraw_notify_text"
                      type="textarea"
                      :placeholder="t('adminSettings.withdrawNotifyTextPlaceholder')"
                      :rows="3"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.withdrawAccountTypes')">
                    <n-input
                      v-model:value="paymentForm.withdraw_account_types_text"
                      type="textarea"
                      :placeholder="t('adminSettings.withdrawAccountTypesPlaceholder')"
                      :rows="3"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingPayment" @click="handleSavePayment">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" :title="t('adminSettings.configDesc')" :bordered="false">
                  <ul style="margin: 0; padding-left: 18px;">
                    <li>{{ t('adminSettings.paymentAlert1') }}</li>
                    <li>{{ t('adminSettings.paymentAlert2') }}</li>
                  </ul>
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="custom" :tab="t('adminSettings.customConfig')">
              <n-space vertical :size="16">
                <n-space justify="end">
                  <TableColumnSelector
                    v-model="customSelectedColumnKeys"
                    :options="customColumnOptions"
                    :visible-count="customVisibleColumnCount"
                    :total-count="customTotalColumnCount"
                    :button-label="t('common.showFields')"
                    :title="t('common.visibleFields')"
                    :hint="t('common.columnVisibilityHint')"
                    :reset-label="t('common.restoreDefaultFields')"
                    @reset="resetCustomSelectedColumns"
                  />
                  <n-button type="primary" @click="showAddModal = true">
                    <template #icon>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" style="width: 1em; height: 1em;">
                        <path d="M11 11V5h2v6h6v2h-6v6h-2v-6H5v-2z" />
                      </svg>
                    </template>
                    {{ t('adminSettings.addConfigItem') }}
                  </n-button>
                </n-space>

                <n-data-table :columns="customVisibleColumns" :data="customSettings" :pagination="false" :bordered="false" :scroll-x="customTableScrollX" />
              </n-space>
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>

        <n-tab-pane name="email-templates" :tab="t('adminSettings.emailTemplates')">
          <EmailTemplates />
        </n-tab-pane>
      </n-tabs>
    </n-spin>

    <n-modal v-model:show="showAddModal" preset="card" :title="t('adminSettings.addConfigTitle')" style="width: 500px;" :mask-closable="false">
      <n-form ref="addFormRef" :model="addForm" :rules="addFormRules" label-placement="left" label-width="100px">
        <n-form-item :label="t('adminSettings.configKey')" path="key">
          <n-input v-model:value="addForm.key" :placeholder="t('adminSettings.configKeyPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configValue')" path="value">
          <n-input v-model:value="addForm.value" :placeholder="t('adminSettings.configValuePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configLabel')" path="label">
          <n-input v-model:value="addForm.label" :placeholder="t('adminSettings.configLabelPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configType')" path="type">
          <n-select v-model:value="addForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configDescription')" path="description">
          <n-input v-model:value="addForm.description" :placeholder="t('adminSettings.configDescriptionPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.isPublic')">
          <n-switch v-model:value="addForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">{{ t('adminSettings.publicConfigHint') }}</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAddModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="adding" @click="handleAddSetting">{{ t('adminSettings.add') }}</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showEditModal" preset="card" :title="t('adminSettings.editConfigTitle')" style="width: 520px;" :mask-closable="false">
      <n-form label-placement="left" label-width="100px">
        <n-form-item :label="t('adminSettings.configKey')">
          <n-input v-model:value="editForm.key" disabled />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configValue')">
          <n-input v-model:value="editForm.value" :placeholder="t('adminSettings.configValuePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configLabel')">
          <n-input v-model:value="editForm.label" :placeholder="t('adminSettings.configLabelPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configType')">
          <n-select v-model:value="editForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configDescription')">
          <n-input v-model:value="editForm.description" :placeholder="t('adminSettings.configDescriptionPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.isPublic')">
          <n-switch v-model:value="editForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">{{ t('adminSettings.publicConfigHint') }}</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="savingEdit" @click="handleSaveSettingEdit">{{ t('common.save') }}</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-card>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
  import { useTableColumnVisibility } from '@/hooks'
  import {
    NAlert,
    NButton,
    NDataTable,
    NDivider,
    NForm,
    NFormItem,
    NInput,
    NInputNumber,
    NModal,
    NSelect,
    NSpace,
    NSpin,
    NSwitch,
    NTabPane,
    NTabs,
    NText,
    type DataTableColumns,
    useMessage,
  } from 'naive-ui'
  import { adminApi } from '@/service/api/admin'
  import EmailTemplates from '@/views/admin/email-templates/index.vue'
  import type { SettingDTO, SettingType } from '@/service/api/admin/settings'
  import { useSettingsStore } from '@/store/settings'
  import { parseBooleanSetting } from '@/utils'

  const message = useMessage()
  const settingsStore = useSettingsStore()
  const { t } = useI18n()

  const loading = ref(true)
  const adding = ref(false)
  const savingEdit = ref(false)
  const showAddModal = ref(false)
  const showEditModal = ref(false)
  const savingBasic = ref(false)
  const savingEmail = ref(false)
  const savingSecurity = ref(false)
  const savingRealnameApi = ref(false)
  const savingPayment = ref(false)
  const testingEmail = ref(false)
  const restartingBackend = ref(false)
  const topTab = ref('system-config')
  const systemSubTab = ref('basic')

  const switchLoading = reactive({
    allow_register: false,
    allow_delete_account: false,
    smtp_ssl: false,
    geetest_enabled: false,
    realname_enabled: false,
    realname_review_required: false,
    realname_api_enabled: false,
    email_verify_enabled: false,
    sms_verify_enabled: false,
    payment_enabled: false,
    withdraw_enabled: false,
  })

  const langOptions = [
    { label: t('adminSettings.langZhCN'), value: 'zhCN' },
    { label: t('adminSettings.langEnUS'), value: 'enUS' },
  ]

const smsProviderOptions = [
  { label: t('adminSettings.smsProviderConsole'), value: 'console' },
  { label: t('adminSettings.smsProviderAliyun'), value: 'aliyun' },
  { label: t('adminSettings.smsProviderTencent'), value: 'tencent' },
  { label: t('adminSettings.smsProviderCustom'), value: 'custom' },
]

const smsBodyFormatOptions = [
  { label: t('adminSettings.formatJSON'), value: 'json' },
  { label: t('adminSettings.formatForm'), value: 'form' },
]

const realnameApiProviderOptions = [
  { label: t('adminSettings.providerAliyun'), value: 'aliyun' },
  { label: t('adminSettings.providerTencent'), value: 'tencent' },
  { label: t('adminSettings.providerBaidu'), value: 'baidu' },
  { label: t('adminSettings.providerCustom'), value: 'custom' },
]

const typeOptions = [
  { label: t('adminSettings.typeString'), value: 'string' },
  { label: t('adminSettings.typeNumber'), value: 'number' },
  { label: t('adminSettings.typeBoolean'), value: 'boolean' },
  { label: t('adminSettings.typeJSON'), value: 'json' },
]

const basicForm = reactive({
  site_name: '',
  site_desc: '',
  site_logo: '',
  copyright: '',
  icp: '',
  version: '',
  default_lang: 'zhCN',
  allow_register: true,
  allow_delete_account: false,
  frontend_url: '',
  backend_api_url: '',
})

const emailForm = reactive({
  email_verify_enabled: true,
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  smtp_ssl: true,
  system_email_name: '',
})

const smsForm = reactive({
  sms_verify_enabled: false,
  sms_provider: 'console',
  sms_access_key: '',
  sms_secret_key: '',
  sms_sign_name: '',
  sms_template_code: '',
  sms_template_code_en: '',
  sms_region: '',
  sms_sdk_app_id: '',
  sms_endpoint: '',
  sms_body_format: 'json',
})

const savingSms = ref(false)

const smsProviderNeedsSignName = computed(() => smsForm.sms_provider !== 'console')
const smsProviderNeedsTemplateCode = computed(() => ['aliyun', 'tencent'].includes(smsForm.sms_provider))
const smsAccessKeyPlaceholder = computed(() => {
  if (smsForm.sms_provider === 'tencent') return t('adminSettings.smsTencentSecretId')
  if (smsForm.sms_provider === 'custom') return t('adminSettings.smsCustomApiKeyOptional')
  return t('adminSettings.smsProviderAccessKey')
})
const smsSecretKeyPlaceholder = computed(() => {
  if (smsForm.sms_provider === 'tencent') return t('adminSettings.smsTencentSecretKey')
  if (smsForm.sms_provider === 'custom') return t('adminSettings.smsCustomApiSecretOptional')
  return t('adminSettings.smsProviderSecretKey')
})
const smsTemplateLabel = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCode') : t('adminSettings.smsTemplateId'))
const smsTemplatePlaceholder = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeExample') : t('adminSettings.smsTemplateIdPlaceholder'))
const smsTemplateEnLabel = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeEnglish') : t('adminSettings.smsTemplateIdEnglish'))
const smsTemplateEnPlaceholder = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeEnglishOptional') : t('adminSettings.smsTemplateIdEnglishOptional'))

const securityForm = reactive({
  geetest_enabled: false,
  geetest_captcha_id: '',
  geetest_captcha_key: '',
  jwt_access_expire: 7200,
  jwt_refresh_expire: 604800,
  login_max_failure: 5,
  login_lock_duration: 10,
  allow_delete_account: false,
  realname_enabled: true,
  realname_review_required: true,
  realname_notify_text: t('adminSettings.realnameNotifyTextDefault'),
})

const realnameApiForm = reactive({
  realname_api_enabled: false,
  realname_api_provider: 'aliyun',
  realname_api_app_key: '',
  realname_api_app_secret: '',
  realname_api_endpoint: '',
})

const paymentForm = reactive({
  payment_enabled: false,
  payment_order_expire_minutes: 30,
  withdraw_enabled: true,
  withdraw_min_amount: 10,
  withdraw_notify_text: t('adminSettings.withdrawNotifyTextDefault'),
  withdraw_account_types_text: '["bank","alipay","wechat","usdt"]',
})

const customSettings = ref<SettingDTO[]>([])

const addFormRef = ref()
const addForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as string,
  description: '',
  is_public: false,
})

const addFormRules = {
  key: [
    { required: true, message: () => t('adminSettings.keyRequired'), trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9_]*$/, message: () => t('adminSettings.keyPattern'), trigger: 'blur' },
  ],
  label: [{ required: true, message: () => t('adminSettings.labelRequired'), trigger: 'blur' }],
}

const editForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as SettingType,
  description: '',
  is_public: false,
})

const customColumns: DataTableColumns<SettingDTO> = [
  { title: t('adminSettings.columnKey'), key: 'key' },
  { title: t('adminSettings.columnLabel'), key: 'label' },
  { title: t('adminSettings.columnValue'), key: 'value', ellipsis: { tooltip: true } },
  { title: t('adminSettings.columnType'), key: 'type', width: 80 },
  {
    title: t('adminSettings.columnPublic'),
    key: 'is_public',
    width: 80,
    render: row => row.is_public ? t('adminSettings.yes') : t('adminSettings.no'),
  },
  {
    title: t('adminSettings.columnActions'),
    key: 'actions',
    width: 180,
    render: (row) => {
      return h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, {
            size: 'small',
            type: 'primary',
            text: true,
            onClick: () => handleEditSetting(row),
          }, () => t('adminSettings.edit')),
          h(NButton, {
            size: 'small',
            type: 'error',
            text: true,
            onClick: () => handleDeleteSetting(row.key),
          }, () => t('adminSettings.delete')),
        ],
      })
    },
  },
]

const customSelectableColumnOptions = computed(() => [
  { key: 'key', label: t('adminSettings.columnKey') },
  { key: 'label', label: t('adminSettings.columnLabel') },
  { key: 'value', label: t('adminSettings.columnValue') },
  { key: 'type', label: t('adminSettings.columnType') },
  { key: 'is_public', label: t('adminSettings.columnPublic') },
])

const {
  columnOptions: customColumnOptions,
  selectedColumnKeys: customSelectedColumnKeys,
  visibleColumns: customVisibleColumns,
  visibleColumnCount: customVisibleColumnCount,
  totalColumnCount: customTotalColumnCount,
  tableScrollX: customTableScrollX,
  resetSelectedColumns: resetCustomSelectedColumns,
} = useTableColumnVisibility<SettingDTO>({
  storageKey: 'admin-settings-custom-list',
  columns: customColumns,
  options: customSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 760,
})

async function loadSettings() {
  loading.value = true
  try {
    const response = await adminApi.settings.list()
    if (response.data?.categories) {
      for (const category of response.data.categories) {
        for (const item of category.items) {
          if (item.key === 'site_name') basicForm.site_name = String(item.value || '')
          if (item.key === 'site_desc') basicForm.site_desc = String(item.value || '')
          if (item.key === 'site_logo') basicForm.site_logo = String(item.value || '')
          if (item.key === 'copyright') basicForm.copyright = String(item.value || '')
          if (item.key === 'icp') basicForm.icp = String(item.value || '')
          if (item.key === 'version') basicForm.version = String(item.value || '')
          if (item.key === 'default_lang') basicForm.default_lang = String(item.value || 'zhCN')

          if (item.key === 'allow_register') basicForm.allow_register = parseBooleanSetting(item.value)
          if (item.key === 'allow_delete_account') securityForm.allow_delete_account = parseBooleanSetting(item.value)
          if (item.key === 'frontend_url') basicForm.frontend_url = String(item.value || '')
          if (item.key === 'backend_api_url') basicForm.backend_api_url = String(item.value || '')

          if (item.key === 'email_verify_enabled') emailForm.email_verify_enabled = parseBooleanSetting(item.value)
          if (item.key === 'smtp_host') emailForm.smtp_host = String(item.value || '')
          if (item.key === 'smtp_port') emailForm.smtp_port = Number(item.value) || 587
          if (item.key === 'smtp_username') emailForm.smtp_username = String(item.value || '')
          if (item.key === 'smtp_password') emailForm.smtp_password = String(item.value || '')
          if (item.key === 'smtp_ssl') emailForm.smtp_ssl = parseBooleanSetting(item.value)
          if (item.key === 'system_email_name') emailForm.system_email_name = String(item.value || '')

          if (item.key === 'sms_verify_enabled') smsForm.sms_verify_enabled = parseBooleanSetting(item.value)
          if (item.key === 'sms_provider') smsForm.sms_provider = String(item.value || 'console')
          if (item.key === 'sms_access_key') smsForm.sms_access_key = String(item.value || '')
          if (item.key === 'sms_secret_key') smsForm.sms_secret_key = String(item.value || '')
          if (item.key === 'sms_sign_name') smsForm.sms_sign_name = String(item.value || '')
          if (item.key === 'sms_template_code') smsForm.sms_template_code = String(item.value || '')
          if (item.key === 'sms_template_code_en') smsForm.sms_template_code_en = String(item.value || '')
          if (item.key === 'sms_region') smsForm.sms_region = String(item.value || '')
          if (item.key === 'sms_sdk_app_id') smsForm.sms_sdk_app_id = String(item.value || '')
          if (item.key === 'sms_endpoint') smsForm.sms_endpoint = String(item.value || '')
          if (item.key === 'sms_body_format') smsForm.sms_body_format = String(item.value || 'json')

          if (item.key === 'geetest_enabled') securityForm.geetest_enabled = parseBooleanSetting(item.value)
          if (item.key === 'geetest_captcha_id') securityForm.geetest_captcha_id = String(item.value || '')
          if (item.key === 'geetest_captcha_key') securityForm.geetest_captcha_key = String(item.value || '')
          if (item.key === 'jwt_access_expire') securityForm.jwt_access_expire = Number(item.value) || 7200
          if (item.key === 'jwt_refresh_expire') securityForm.jwt_refresh_expire = Number(item.value) || 604800
          if (item.key === 'login_max_failure') securityForm.login_max_failure = Number(item.value) || 5
          if (item.key === 'login_lock_duration') securityForm.login_lock_duration = Number(item.value) || 10
          if (item.key === 'realname_enabled') securityForm.realname_enabled = parseBooleanSetting(item.value)
          if (item.key === 'realname_review_required') securityForm.realname_review_required = parseBooleanSetting(item.value)
          if (item.key === 'realname_notify_text') securityForm.realname_notify_text = String(item.value || t('adminSettings.realnameNotifyTextDefault'))

          if (item.key === 'realname_api_enabled') realnameApiForm.realname_api_enabled = parseBooleanSetting(item.value)
          if (item.key === 'realname_api_provider') realnameApiForm.realname_api_provider = String(item.value || 'aliyun')
          if (item.key === 'realname_api_app_key') realnameApiForm.realname_api_app_key = String(item.value || '')
          if (item.key === 'realname_api_app_secret') realnameApiForm.realname_api_app_secret = String(item.value || '')
          if (item.key === 'realname_api_endpoint') realnameApiForm.realname_api_endpoint = String(item.value || '')

          if (item.key === 'payment_enabled') paymentForm.payment_enabled = parseBooleanSetting(item.value)
          if (item.key === 'payment_order_expire_minutes') paymentForm.payment_order_expire_minutes = Number(item.value) || 30
          if (item.key === 'withdraw_enabled') paymentForm.withdraw_enabled = parseBooleanSetting(item.value)
          if (item.key === 'withdraw_min_amount') paymentForm.withdraw_min_amount = Number(item.value) || 10
          if (item.key === 'withdraw_notify_text') paymentForm.withdraw_notify_text = String(item.value || t('adminSettings.withdrawNotifyTextDefault'))
          if (item.key === 'withdraw_account_types') paymentForm.withdraw_account_types_text = typeof item.value === 'string' ? item.value : JSON.stringify(item.value || ['bank', 'alipay', 'wechat', 'usdt'])
        }

        if (category.category === 'custom') {
          customSettings.value = category.items
        }
      }
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.loadSettingsFailed'))
  }
  finally {
    loading.value = false
  }
}

async function handleUpdateAllowRegister(nextValue: boolean) {
  const prev = basicForm.allow_register
  basicForm.allow_register = nextValue
  switchLoading.allow_register = true
  try {
    const res = await adminApi.settings.update('allow_register', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ allow_register: nextValue })
      message.success(res.message || t('adminSettings.registerSwitchUpdated'))
    }
    else {
      basicForm.allow_register = prev
      message.error(res.message || t('adminSettings.registerSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    basicForm.allow_register = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.allow_register = false
  }
}

async function handleUpdateAllowDeleteAccount(nextValue: boolean) {
  const prev = securityForm.allow_delete_account
  securityForm.allow_delete_account = nextValue
  switchLoading.allow_delete_account = true
  try {
    const res = await adminApi.settings.update('allow_delete_account', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ allow_delete_account: nextValue })
      message.success(res.message || t('adminSettings.deleteAccountSwitchUpdated'))
    }
    else {
      securityForm.allow_delete_account = prev
      message.error(res.message || t('adminSettings.deleteAccountSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.allow_delete_account = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.allow_delete_account = false
  }
}

async function handleUpdateSmtpSSL(nextValue: boolean) {
  const prev = emailForm.smtp_ssl
  emailForm.smtp_ssl = nextValue
  switchLoading.smtp_ssl = true
  try {
    const res = await adminApi.settings.update('smtp_ssl', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.smtpSslUpdated'))
    }
    else {
      emailForm.smtp_ssl = prev
      message.error(res.message || t('adminSettings.smtpSslUpdateFailed'))
    }
  }
  catch (error: any) {
    emailForm.smtp_ssl = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.smtp_ssl = false
  }
}

async function handleUpdateEmailVerifyEnabled(nextValue: boolean) {
  const prev = emailForm.email_verify_enabled
  emailForm.email_verify_enabled = nextValue
  switchLoading.email_verify_enabled = true
  try {
    const res = await adminApi.settings.update('email_verify_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ email_verify_enabled: nextValue })
      message.success(res.message || t('adminSettings.emailVerifySwitchUpdated'))
    }
    else {
      emailForm.email_verify_enabled = prev
      message.error(res.message || t('adminSettings.emailVerifySwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    emailForm.email_verify_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.email_verify_enabled = false
  }
}

async function handleUpdateSmsVerifyEnabled(nextValue: boolean) {
  const prev = smsForm.sms_verify_enabled
  smsForm.sms_verify_enabled = nextValue
  switchLoading.sms_verify_enabled = true
  try {
    const res = await adminApi.settings.update('sms_verify_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ sms_verify_enabled: nextValue })
      message.success(res.message || t('adminSettings.smsVerifySwitchUpdated'))
    }
    else {
      smsForm.sms_verify_enabled = prev
      message.error(res.message || t('adminSettings.smsVerifySwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    smsForm.sms_verify_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.sms_verify_enabled = false
  }
}

async function handleSaveSms() {
  savingSms.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      sms_provider: smsForm.sms_provider,
      sms_access_key: smsForm.sms_access_key,
      sms_secret_key: smsForm.sms_secret_key,
      sms_sign_name: smsForm.sms_sign_name,
      sms_template_code: smsForm.sms_template_code,
      sms_template_code_en: smsForm.sms_template_code_en,
      sms_region: smsForm.sms_region,
      sms_sdk_app_id: smsForm.sms_sdk_app_id,
      sms_endpoint: smsForm.sms_endpoint,
      sms_body_format: smsForm.sms_body_format,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.smsSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.smsSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingSms.value = false
  }
}

async function handleUpdateGeetestEnabled(nextValue: boolean) {
  const prev = securityForm.geetest_enabled
  securityForm.geetest_enabled = nextValue
  switchLoading.geetest_enabled = true
  try {
    const res = await adminApi.settings.update('geetest_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ geetest_enabled: nextValue })
      message.success(res.message || t('adminSettings.geetestSwitchUpdated'))
    }
    else {
      securityForm.geetest_enabled = prev
      message.error(res.message || t('adminSettings.geetestSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.geetest_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.geetest_enabled = false
  }
}

async function handleUpdateRealnameEnabled(nextValue: boolean) {
  const prev = securityForm.realname_enabled
  securityForm.realname_enabled = nextValue
  switchLoading.realname_enabled = true
  try {
    const res = await adminApi.settings.update('realname_enabled', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameSwitchUpdated'))
    }
    else {
      securityForm.realname_enabled = prev
      message.error(res.message || t('adminSettings.realnameSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.realname_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.realname_enabled = false
  }
}

async function handleUpdateRealnameReviewRequired(nextValue: boolean) {
  const prev = securityForm.realname_review_required
  securityForm.realname_review_required = nextValue
  switchLoading.realname_review_required = true
  try {
    const res = await adminApi.settings.update('realname_review_required', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameReviewSwitchUpdated'))
    }
    else {
      securityForm.realname_review_required = prev
      message.error(res.message || t('adminSettings.realnameReviewSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.realname_review_required = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.realname_review_required = false
  }
}

async function handleUpdateRealnameApiEnabled(nextValue: boolean) {
  const prev = realnameApiForm.realname_api_enabled
  realnameApiForm.realname_api_enabled = nextValue
  switchLoading.realname_api_enabled = true
  try {
    const res = await adminApi.settings.update('realname_api_enabled', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameApiSwitchUpdated'))
    }
    else {
      realnameApiForm.realname_api_enabled = prev
      message.error(res.message || t('adminSettings.realnameApiSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    realnameApiForm.realname_api_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.realname_api_enabled = false
  }
}

async function handleSaveRealnameApi() {
  savingRealnameApi.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      realname_api_provider: realnameApiForm.realname_api_provider,
      realname_api_app_key: realnameApiForm.realname_api_app_key,
      realname_api_app_secret: realnameApiForm.realname_api_app_secret,
      realname_api_endpoint: realnameApiForm.realname_api_endpoint,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameApiSaveSuccess'))
    }
    else {
      message.error(res.message || t('adminSettings.realnameApiSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingRealnameApi.value = false
  }
}

async function handleUpdatePaymentEnabled(nextValue: boolean) {
  const prev = paymentForm.payment_enabled
  paymentForm.payment_enabled = nextValue
  switchLoading.payment_enabled = true
  try {
    const res = await adminApi.settings.update('payment_enabled', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.paymentSwitchUpdated'))
    }
    else {
      paymentForm.payment_enabled = prev
      message.error(res.message || t('adminSettings.paymentSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    paymentForm.payment_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.payment_enabled = false
  }
}

async function handleUpdateWithdrawEnabled(nextValue: boolean) {
  const prev = paymentForm.withdraw_enabled
  paymentForm.withdraw_enabled = nextValue
  switchLoading.withdraw_enabled = true
  try {
    const res = await adminApi.settings.update('withdraw_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ withdraw_enabled: nextValue })
      message.success(res.message || t('adminSettings.withdrawSwitchUpdated'))
    }
    else {
      paymentForm.withdraw_enabled = prev
      message.error(res.message || t('adminSettings.withdrawSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    paymentForm.withdraw_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.withdraw_enabled = false
  }
}

async function handleSavePayment() {
  savingPayment.value = true
  try {
    const parsedAccountTypes = JSON.parse(paymentForm.withdraw_account_types_text || '[]')
    if (!Array.isArray(parsedAccountTypes) || parsedAccountTypes.length === 0 || parsedAccountTypes.some(item => typeof item !== 'string' || !item.trim())) {
      throw new Error(t('adminSettings.invalidAccountTypes'))
    }
    const res = await adminApi.settings.batchUpdate({
      payment_order_expire_minutes: String(paymentForm.payment_order_expire_minutes),
      withdraw_min_amount: String(paymentForm.withdraw_min_amount),
      withdraw_notify_text: paymentForm.withdraw_notify_text,
      withdraw_account_types: paymentForm.withdraw_account_types_text,
    })
    if (res.isSuccess) {
      settingsStore.updateConfig({
        withdraw_min_amount: paymentForm.withdraw_min_amount,
        withdraw_notify_text: paymentForm.withdraw_notify_text,
        withdraw_account_types: parsedAccountTypes,
      })
      message.success(res.message || t('adminSettings.paymentSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.paymentSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingPayment.value = false
  }
}

async function handleSaveBasic() {
  savingBasic.value = true
  try {
    const frontendUrl = basicForm.frontend_url.trim().replace(/\/+$/, '')
    const backendApiUrl = basicForm.backend_api_url.trim().replace(/\/+$/, '')
    basicForm.frontend_url = frontendUrl
    basicForm.backend_api_url = backendApiUrl
    const res = await adminApi.settings.batchUpdate({
      site_name: basicForm.site_name,
      site_desc: basicForm.site_desc,
      site_logo: basicForm.site_logo,
      copyright: basicForm.copyright,
      icp: basicForm.icp,
      version: basicForm.version,
      default_lang: basicForm.default_lang,
      frontend_url: frontendUrl,
      backend_api_url: backendApiUrl,
    })
    if (res.isSuccess) {
      settingsStore.updateConfig({
        site_name: basicForm.site_name,
        site_desc: basicForm.site_desc,
        site_logo: basicForm.site_logo,
        copyright: basicForm.copyright,
        icp: basicForm.icp,
        version: basicForm.version,
        default_lang: basicForm.default_lang,
      })
      message.success(res.message || t('adminSettings.basicSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.basicSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingBasic.value = false
  }
}

async function handleSaveEmail() {
  savingEmail.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      smtp_host: emailForm.smtp_host,
      smtp_port: String(emailForm.smtp_port),
      smtp_username: emailForm.smtp_username,
      smtp_password: emailForm.smtp_password,
      system_email_name: emailForm.system_email_name,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.emailSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.emailSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingEmail.value = false
  }
}

async function handleTestEmail() {
  testingEmail.value = true
  try {
    const res = await adminApi.emailTemplate.sendTest({
      to: emailForm.smtp_username,
    })
    if (res.isSuccess) {
      message.success(res.data?.message || t('adminSettings.testEmailSent'))
    }
    else {
      message.error(res.message || t('adminSettings.testEmailFailed'))
    }
  }
  catch (error: any) {
    message.error(error?.message || t('adminSettings.testEmailFailed'))
  }
  finally {
    testingEmail.value = false
  }
}

async function handleSaveSecurity() {
  savingSecurity.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      geetest_captcha_id: securityForm.geetest_captcha_id,
      geetest_captcha_key: securityForm.geetest_captcha_key,
      jwt_access_expire: String(securityForm.jwt_access_expire),
      jwt_refresh_expire: String(securityForm.jwt_refresh_expire),
      login_max_failure: String(securityForm.login_max_failure),
      login_lock_duration: String(securityForm.login_lock_duration),
      realname_notify_text: securityForm.realname_notify_text,
    })
    if (res.isSuccess) {
      settingsStore.updateConfig({
        geetest_captcha_id: securityForm.geetest_captcha_id,
      })
      message.success(res.message || t('adminSettings.securitySettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.securitySettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingSecurity.value = false
  }
}

async function handleRestartBackend() {
  restartingBackend.value = true
  try {
    const res = await adminApi.settings.restartBackend()
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.restartBackendRequested'))
    }
    else {
      message.error(res.message || t('adminSettings.restartBackendFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.restartFailed'))
  }
  finally {
    restartingBackend.value = false
  }
}

async function handleAddSetting() {
  try {
    await addFormRef.value?.validate()
  }
  catch {
    return
  }

  adding.value = true
  try {
    const res = await adminApi.settings.create({
      key: addForm.key,
      value: addForm.value,
      type: addForm.type as SettingType,
      category: 'custom',
      label: addForm.label,
      description: addForm.description,
      is_public: addForm.is_public,
      is_editable: true,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.configItemAdded'))
      showAddModal.value = false
      addForm.key = ''
      addForm.value = ''
      addForm.label = ''
      addForm.type = 'string'
      addForm.description = ''
      addForm.is_public = false
      await loadSettings()
    }
    else {
      message.error(res.message || t('adminSettings.configItemAddFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.addFailed'))
  }
  finally {
    adding.value = false
  }
}

async function handleDeleteSetting(key: string) {
  try {
    const res = await adminApi.settings.delete(key)
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.configItemDeleted'))
      await loadSettings()
    }
    else {
      message.error(res.message || t('adminSettings.configItemDeleteFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.deleteFailed') + (error.message || ''))
  }
}

function handleEditSetting(row: SettingDTO) {
  editForm.key = row.key
  editForm.value = row.value == null ? '' : String(row.value)
  editForm.label = row.label || ''
  editForm.type = row.type
  editForm.description = row.description || ''
  editForm.is_public = Boolean(row.is_public)
  showEditModal.value = true
}

async function handleSaveSettingEdit() {
  if (!editForm.key) {
    return
  }
  savingEdit.value = true
  try {
    const res = await adminApi.settings.updateMeta(editForm.key, {
      value: editForm.value,
      type: editForm.type,
      category: 'custom',
      label: editForm.label,
      description: editForm.description,
      is_public: editForm.is_public,
      is_editable: true,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.configItemUpdated'))
      showEditModal.value = false
      await loadSettings()
    }
    else {
      message.error(res.message || t('adminSettings.configItemUpdateFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.editFailed'))
  }
  finally {
    savingEdit.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped></style>
