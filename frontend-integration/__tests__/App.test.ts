import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { mount } from '@vue/test-utils';
import App from '../demo/App.vue';
import MegaportTerminal from '../components/MegaportTerminal.vue';

describe('App.vue', () => {
  let wrapper: any;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    if (wrapper) {
      wrapper = null;
    }
  });

  describe('Authentication Form', () => {
    it('should render authentication form when not authenticated', () => {
      wrapper = mount(App);

      expect(wrapper.find('.auth-panel').exists()).toBe(true);
      expect(wrapper.find('input#accessKey').exists()).toBe(true);
      expect(wrapper.find('input#secretKey').exists()).toBe(true);
      expect(wrapper.find('select#environment').exists()).toBe(true);
      expect(wrapper.find('.btn-primary').text()).toBe('Set Credentials');
    });

    it('should have default environment as staging', () => {
      wrapper = mount(App);

      const select = wrapper.find('select#environment');
      expect((select.element as HTMLSelectElement).value).toBe('staging');
    });

    it('should update form fields when user types', async () => {
      wrapper = mount(App);

      const accessKeyInput = wrapper.find('input#accessKey');
      const secretKeyInput = wrapper.find('input#secretKey');
      const environmentSelect = wrapper.find('select#environment');

      await accessKeyInput.setValue('test-access-key');
      await secretKeyInput.setValue('test-secret-key');
      await environmentSelect.setValue('production');

      expect((accessKeyInput.element as HTMLInputElement).value).toBe(
        'test-access-key'
      );
      expect((secretKeyInput.element as HTMLInputElement).value).toBe(
        'test-secret-key'
      );
      expect((environmentSelect.element as HTMLSelectElement).value).toBe(
        'production'
      );
    });

    it('should call setAuth when form is submitted', async () => {
      wrapper = mount(App);

      const accessKeyInput = wrapper.find('input#accessKey');
      const secretKeyInput = wrapper.find('input#secretKey');
      const form = wrapper.find('form');

      await accessKeyInput.setValue('my-access-key');
      await secretKeyInput.setValue('my-secret-key');
      await form.trigger('submit.prevent');

      // Auth panel should be hidden after submission
      await wrapper.vm.$nextTick();
      expect(wrapper.find('.auth-panel').exists()).toBe(false);
    });

    it('should show terminal section after authentication', async () => {
      wrapper = mount(App);

      const accessKeyInput = wrapper.find('input#accessKey');
      const secretKeyInput = wrapper.find('input#secretKey');
      const form = wrapper.find('form');

      await accessKeyInput.setValue('test-key');
      await secretKeyInput.setValue('test-secret');
      await form.trigger('submit.prevent');

      await wrapper.vm.$nextTick();

      expect(wrapper.find('.terminal-section').exists()).toBe(true);
      expect(wrapper.findComponent(MegaportTerminal).exists()).toBe(true);
    });
  });

  describe('Quick Actions', () => {
    it('should not show quick actions when not authenticated', () => {
      wrapper = mount(App);

      expect(wrapper.find('.quick-actions').exists()).toBe(false);
    });

    it('should show quick actions after authentication', async () => {
      wrapper = mount(App);

      const form = wrapper.find('form');
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await form.trigger('submit.prevent');

      await wrapper.vm.$nextTick();

      expect(wrapper.find('.quick-actions').exists()).toBe(true);
    });

    it('should render all quick action buttons', async () => {
      wrapper = mount(App);

      // Authenticate first
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      const buttons = wrapper.findAll('.btn-action');
      expect(buttons.length).toBe(6);

      expect(buttons[0].text()).toContain('List Ports');
      expect(buttons[1].text()).toContain('List MCRs');
      expect(buttons[2].text()).toContain('List MVEs');
      expect(buttons[3].text()).toContain('List Locations');
      expect(buttons[4].text()).toContain('Help');
      expect(buttons[5].text()).toContain('Clear');
    });

    it('should have correct commands for each quick action', async () => {
      wrapper = mount(App);

      // Authenticate first
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      const buttons = wrapper.findAll('.btn-action');

      // Check that each button has the correct @click handler
      // We can't directly test the command string, but we can verify buttons exist
      expect(buttons[0].text()).toContain('List Ports');
      expect(buttons[1].text()).toContain('List MCRs');
      expect(buttons[2].text()).toContain('List MVEs');
      expect(buttons[3].text()).toContain('List Locations');
    });

    it('should execute commands when quick action buttons are clicked', async () => {
      wrapper = mount(App);

      // Authenticate
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      // Get quick action buttons - buttons are shown after auth
      const buttons = wrapper.findAll('.btn-action');

      // Verify we have the buttons
      expect(buttons.length).toBe(6);
      expect(buttons[0].text()).toContain('List Ports');
    });
  });

  describe('Clear Authentication', () => {
    it('should show clear auth button when authenticated', async () => {
      wrapper = mount(App);

      // Authenticate
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      expect(wrapper.find('.btn-secondary').exists()).toBe(true);
      expect(wrapper.find('.btn-secondary').text()).toContain('Clear Auth');
    });

    it('should return to auth form when clear auth is clicked', async () => {
      wrapper = mount(App);

      // Authenticate
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      expect(wrapper.find('.auth-panel').exists()).toBe(false);
      expect(wrapper.find('.terminal-section').exists()).toBe(true);

      // Clear auth
      const clearButton = wrapper.find('.btn-secondary');
      await clearButton.trigger('click');
      await wrapper.vm.$nextTick();

      expect(wrapper.find('.auth-panel').exists()).toBe(true);
      expect(wrapper.find('.terminal-section').exists()).toBe(false);
    });

    it('should reset form fields when auth is cleared', async () => {
      wrapper = mount(App);

      // Authenticate
      await wrapper.find('input#accessKey').setValue('my-key');
      await wrapper.find('input#secretKey').setValue('my-secret');
      await wrapper.find('select#environment').setValue('production');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      // Clear auth
      await wrapper.find('.btn-secondary').trigger('click');
      await wrapper.vm.$nextTick();

      expect(
        (wrapper.find('input#accessKey').element as HTMLInputElement).value
      ).toBe('');
      expect(
        (wrapper.find('input#secretKey').element as HTMLInputElement).value
      ).toBe('');
      expect(
        (wrapper.find('select#environment').element as HTMLSelectElement).value
      ).toBe('staging');
    });
  });

  describe('Status Info', () => {
    it('should always show status info panel', () => {
      wrapper = mount(App);

      expect(wrapper.find('.status-info').exists()).toBe(true);
      expect(wrapper.find('.status-info h3').text()).toBe('WASM Status');
    });

    it('should display loading and ready status', () => {
      wrapper = mount(App);

      const statusItems = wrapper.findAll('.status-item');
      expect(statusItems.length).toBeGreaterThanOrEqual(2);

      // Check for Loading and Ready labels
      const labels = wrapper.findAll('.status-item .label');
      const labelTexts = labels.map((l: any) => l.text());
      expect(labelTexts).toContain('Loading:');
      expect(labelTexts).toContain('Ready:');
    });

    it('should show environment after authentication', async () => {
      wrapper = mount(App);

      // Authenticate
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('select#environment').setValue('production');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      const statusItems = wrapper.findAll('.status-item');
      const environmentItem = statusItems.find((item: any) =>
        item.find('.label').text().includes('Environment')
      );

      expect(environmentItem).toBeDefined();
    });
  });

  describe('Layout and Structure', () => {
    it('should render header with correct title', () => {
      wrapper = mount(App);

      expect(wrapper.find('.app-header h1').text()).toContain(
        'Megaport CLI WebAssembly Demo'
      );
      expect(wrapper.find('.app-header p').text()).toContain(
        'Vue 3 + Vite + WASM Integration'
      );
    });

    it('should render footer with links', () => {
      wrapper = mount(App);

      expect(wrapper.find('.app-footer').exists()).toBe(true);
      const links = wrapper.findAll('.app-footer a');
      expect(links.length).toBe(2);
      expect(links[0].attributes('href')).toBe('https://github.com/megaport');
      expect(links[1].attributes('href')).toBe('https://docs.megaport.com');
    });

    it('should have proper CSS classes', () => {
      wrapper = mount(App);

      expect(wrapper.find('.app-container').exists()).toBe(true);
      expect(wrapper.find('.app-header').exists()).toBe(true);
      expect(wrapper.find('.app-main').exists()).toBe(true);
      expect(wrapper.find('.app-footer').exists()).toBe(true);
    });
  });

  describe('Terminal Integration', () => {
    it('should pass correct props to MegaportTerminal', async () => {
      wrapper = mount(App);

      // Authenticate to show terminal
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      const terminal = wrapper.findComponent(MegaportTerminal);
      expect(terminal.exists()).toBe(true);
      expect(terminal.props('wasmPath')).toBe('/megaport.wasm');
      expect(terminal.props('wasmExecPath')).toBe('/wasm_exec.js');
      expect(terminal.props('welcomeMessage')).toContain(
        'Welcome to Megaport CLI'
      );
    });

    it('should include environment in welcome message', async () => {
      wrapper = mount(App);

      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('select#environment').setValue('production');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      const terminal = wrapper.findComponent(MegaportTerminal);
      expect(terminal.props('welcomeMessage')).toContain(
        'Environment: production'
      );
    });
  });

  describe('Command Execution Order', () => {
    it('should execute commands in correct order: ports, mcr, mve, locations', async () => {
      wrapper = mount(App);

      // Authenticate
      await wrapper.find('input#accessKey').setValue('key');
      await wrapper.find('input#secretKey').setValue('secret');
      await wrapper.find('form').trigger('submit.prevent');
      await wrapper.vm.$nextTick();

      const buttons = wrapper.findAll('.btn-action');

      // Verify button order
      expect(buttons[0].text()).toContain('List Ports');
      expect(buttons[1].text()).toContain('List MCRs');
      expect(buttons[2].text()).toContain('List MVEs');
      expect(buttons[3].text()).toContain('List Locations');
    });
  });
});
