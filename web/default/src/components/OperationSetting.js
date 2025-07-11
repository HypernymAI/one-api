import React, { useEffect, useState } from 'react';
import { Divider, Form, Grid, Header } from 'semantic-ui-react';
import { API, showError, showSuccess, timestamp2string, verifyJSON } from '../helpers';

const OperationSetting = () => {
  let now = new Date();
  let [inputs, setInputs] = useState({
    QuotaForNewUser: 0,
    QuotaForInviter: 0,
    QuotaForInvitee: 0,
    QuotaRemindThreshold: 0,
    PreConsumedQuota: 0,
    ModelRatio: '',
    CompletionRatio: '',
    GroupRatio: '',
    TopUpLink: '',
    ChatLink: '',
    QuotaPerUnit: 0,
    AutomaticDisableChannelEnabled: '',
    AutomaticEnableChannelEnabled: '',
    ChannelDisableThreshold: 0,
    LogConsumeEnabled: '',
    DisplayInCurrencyEnabled: '',
    DisplayTokenStatEnabled: '',
    ApproximateTokenEnabled: '',
    RetryTimes: 0
  });
  const [originInputs, setOriginInputs] = useState({});
  let [loading, setLoading] = useState(false);
  let [historyTimestamp, setHistoryTimestamp] = useState(timestamp2string(now.getTime() / 1000 - 30 * 24 * 3600)); // a month ago

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (item.key === 'ModelRatio' || item.key === 'GroupRatio' || item.key === 'CompletionRatio') {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }
        if (item.value === '{}') {
          item.value = '';
        }
        newInputs[item.key] = item.value;
      });
      setInputs(newInputs);
      setOriginInputs(newInputs);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions().then();
  }, []);

  const updateOption = async (key, value) => {
    setLoading(true);
    if (key.endsWith('Enabled')) {
      value = inputs[key] === 'true' ? 'false' : 'true';
    }
    const res = await API.put('/api/option/', {
      key,
      value
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, [key]: value }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (e, { name, value }) => {
    if (name.endsWith('Enabled')) {
      await updateOption(name, value);
    } else {
      setInputs((inputs) => ({ ...inputs, [name]: value }));
    }
  };

  const submitConfig = async (group) => {
    switch (group) {
      case 'monitor':
        if (originInputs['ChannelDisableThreshold'] !== inputs.ChannelDisableThreshold) {
          await updateOption('ChannelDisableThreshold', inputs.ChannelDisableThreshold);
        }
        if (originInputs['QuotaRemindThreshold'] !== inputs.QuotaRemindThreshold) {
          await updateOption('QuotaRemindThreshold', inputs.QuotaRemindThreshold);
        }
        break;
      case 'ratio':
        if (originInputs['ModelRatio'] !== inputs.ModelRatio) {
          if (!verifyJSON(inputs.ModelRatio)) {
            showError('Model ratio is not a valid JSON string');
            return;
          }
          await updateOption('ModelRatio', inputs.ModelRatio);
        }
        if (originInputs['GroupRatio'] !== inputs.GroupRatio) {
          if (!verifyJSON(inputs.GroupRatio)) {
            showError('Group ratio is not a valid JSON string');
            return;
          }
          await updateOption('GroupRatio', inputs.GroupRatio);
        }
        if (originInputs['CompletionRatio'] !== inputs.CompletionRatio) {
          if (!verifyJSON(inputs.CompletionRatio)) {
            showError('Completion ratio is not a valid JSON string');
            return;
          }
          await updateOption('CompletionRatio', inputs.CompletionRatio);
        }
        break;
      case 'quota':
        if (originInputs['QuotaForNewUser'] !== inputs.QuotaForNewUser) {
          await updateOption('QuotaForNewUser', inputs.QuotaForNewUser);
        }
        if (originInputs['QuotaForInvitee'] !== inputs.QuotaForInvitee) {
          await updateOption('QuotaForInvitee', inputs.QuotaForInvitee);
        }
        if (originInputs['QuotaForInviter'] !== inputs.QuotaForInviter) {
          await updateOption('QuotaForInviter', inputs.QuotaForInviter);
        }
        if (originInputs['PreConsumedQuota'] !== inputs.PreConsumedQuota) {
          await updateOption('PreConsumedQuota', inputs.PreConsumedQuota);
        }
        break;
      case 'general':
        if (originInputs['TopUpLink'] !== inputs.TopUpLink) {
          await updateOption('TopUpLink', inputs.TopUpLink);
        }
        if (originInputs['ChatLink'] !== inputs.ChatLink) {
          await updateOption('ChatLink', inputs.ChatLink);
        }
        if (originInputs['QuotaPerUnit'] !== inputs.QuotaPerUnit) {
          await updateOption('QuotaPerUnit', inputs.QuotaPerUnit);
        }
        if (originInputs['RetryTimes'] !== inputs.RetryTimes) {
          await updateOption('RetryTimes', inputs.RetryTimes);
        }
        break;
    }
  };

  const deleteHistoryLogs = async () => {
    console.log(inputs);
    const res = await API.delete(`/api/log/?target_timestamp=${Date.parse(historyTimestamp) / 1000}`);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(`${data} logs have been cleaned!`);
      return;
    }
    showError('Log cleanup failed: ' + message);
  };

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading}>
          <Header as='h3'>
            General Settings
          </Header>
          <Form.Group widths={4}>
            <Form.Input
              label='Top-up Link'
              name='TopUpLink'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.TopUpLink}
              type='link'
              placeholder='e.g. card issuing website purchase link'
            />
            <Form.Input
              label='Chat Page Link'
              name='ChatLink'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.ChatLink}
              type='link'
              placeholder='e.g. ChatGPT Next Web deployment address'
            />
            <Form.Input
              label='Quota per Unit Dollar'
              name='QuotaPerUnit'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaPerUnit}
              type='number'
              step='0.01'
              placeholder='Quota that can be exchanged for one unit of currency'
            />
            <Form.Input
              label='Retry Times on Failure'
              name='RetryTimes'
              type={'number'}
              step='1'
              min='0'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.RetryTimes}
              placeholder='Number of retry attempts on failure'
            />
          </Form.Group>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.DisplayInCurrencyEnabled === 'true'}
              label='Display quota in currency form'
              name='DisplayInCurrencyEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.DisplayTokenStatEnabled === 'true'}
              label='Billing related API displays token quota instead of user quota'
              name='DisplayTokenStatEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.ApproximateTokenEnabled === 'true'}
              label='Use approximate method to estimate token count to reduce computation'
              name='ApproximateTokenEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={() => {
            submitConfig('general').then();
          }}>Save General Settings</Form.Button>
          <Divider />
          <Header as='h3'>
            Log Settings
          </Header>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.LogConsumeEnabled === 'true'}
              label='Enable quota consumption logging'
              name='LogConsumeEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths={4}>
            <Form.Input label='Target Time' value={historyTimestamp} type='datetime-local'
                        name='history_timestamp'
                        onChange={(e, { name, value }) => {
                          setHistoryTimestamp(value);
                        }} />
          </Form.Group>
          <Form.Button onClick={() => {
            deleteHistoryLogs().then();
          }}>Clean History Logs</Form.Button>
          <Divider />
          <Header as='h3'>
            Monitoring Settings
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label='Maximum Response Time'
              name='ChannelDisableThreshold'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.ChannelDisableThreshold}
              type='number'
              min='0'
              placeholder='In seconds, channel will be automatically disabled if exceeded during full test'
            />
            <Form.Input
              label='Quota Reminder Threshold'
              name='QuotaRemindThreshold'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaRemindThreshold}
              type='number'
              min='0'
              placeholder='Email reminder will be sent when quota falls below this threshold'
            />
          </Form.Group>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.AutomaticDisableChannelEnabled === 'true'}
              label='Automatically disable channel on failure'
              name='AutomaticDisableChannelEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.AutomaticEnableChannelEnabled === 'true'}
              label='Automatically enable channel on success'
              name='AutomaticEnableChannelEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={() => {
            submitConfig('monitor').then();
          }}>Save Monitoring Settings</Form.Button>
          <Divider />
          <Header as='h3'>
            Quota Settings
          </Header>
          <Form.Group widths={4}>
            <Form.Input
              label='Initial Quota for New Users'
              name='QuotaForNewUser'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaForNewUser}
              type='number'
              min='0'
              placeholder='e.g.: 100'
            />
            <Form.Input
              label='Pre-deducted Request Quota'
              name='PreConsumedQuota'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.PreConsumedQuota}
              type='number'
              min='0'
              placeholder='Refund excess or charge shortage after request'
            />
            <Form.Input
              label='Reward Quota for Inviting New Users'
              name='QuotaForInviter'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaForInviter}
              type='number'
              min='0'
              placeholder='e.g.: 2000'
            />
            <Form.Input
              label='Reward Quota for New Users Using Invitation Code'
              name='QuotaForInvitee'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.QuotaForInvitee}
              type='number'
              min='0'
              placeholder='e.g.: 1000'
            />
          </Form.Group>
          <Form.Button onClick={() => {
            submitConfig('quota').then();
          }}>Save Quota Settings</Form.Button>
          <Divider />
          <Header as='h3'>
            Ratio Settings
          </Header>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='Model Ratio'
              name='ModelRatio'
              onChange={handleInputChange}
              style={{ minHeight: 250, fontFamily: 'JetBrains Mono, Consolas' }}
              autoComplete='new-password'
              value={inputs.ModelRatio}
              placeholder='A JSON text with model names as keys and ratios as values'
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='Completion Ratio'
              name='CompletionRatio'
              onChange={handleInputChange}
              style={{ minHeight: 250, fontFamily: 'JetBrains Mono, Consolas' }}
              autoComplete='new-password'
              value={inputs.CompletionRatio}
              placeholder='A JSON text with model names as keys and ratios as values. This ratio setting is the ratio of model completion rate to prompt rate, which can forcibly override the internal ratio of One API'
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='Group Ratio'
              name='GroupRatio'
              onChange={handleInputChange}
              style={{ minHeight: 250, fontFamily: 'JetBrains Mono, Consolas' }}
              autoComplete='new-password'
              value={inputs.GroupRatio}
              placeholder='A JSON text with group names as keys and ratios as values'
            />
          </Form.Group>
          <Form.Button onClick={() => {
            submitConfig('ratio').then();
          }}>Save Ratio Settings</Form.Button>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default OperationSetting;
