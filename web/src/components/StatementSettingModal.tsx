import React, { useEffect, useState } from 'react';
import { Modal, Form, message, Spin, InputNumber } from 'antd';
import {
  IStatementConfig,
  getInstanceStatementConfig,
  updateInstanceStatementConfig,
} from '@/services/statement';

const styles = require('./ConfigInstanceModal.less');

interface Props {
  visible: boolean;
  onClose: () => void;

  instanceId: string;
}

const formItemLayout = {
  labelCol: {
    xs: { span: 24 },
    sm: { span: 8 },
  },
  wrapperCol: {
    xs: { span: 24 },
    sm: { span: 16 },
  },
};

function StatementSettingModal({ visible, onClose, instanceId }: Props) {
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [config, setConfig] = useState<IStatementConfig | null>(null);

  useEffect(() => {
    async function fetchConfig() {
      if (instanceId === '') {
        return;
      }

      setLoading(true);
      const res = await getInstanceStatementConfig(instanceId);
      if (res !== undefined) {
        setConfig(res as IStatementConfig);
      }
      setLoading(false);
    }
    fetchConfig();
  }, [instanceId]);

  async function submit() {
    setSubmitting(true);
    const res = await updateInstanceStatementConfig(instanceId, config as IStatementConfig);
    setSubmitting(false);
    if (res !== undefined) {
      onClose();
      message.success(`${instanceId} 设置 statement 成功！`);
    }
  }

  function handleConfigChange(configKey: string, configValue: number | undefined) {
    setConfig({
      ...(config as IStatementConfig),
      [configKey]: configValue,
    });
  }

  return (
    <Modal
      visible={visible}
      onCancel={onClose}
      onOk={submit}
      title="设置"
      confirmLoading={submitting}
      okButtonProps={{ disabled: loading || config === null }}
    >
      {loading && <Spin size="small" style={{ marginLeft: 8, marginRight: 8 }} />}
      {!loading && config && (
        <Form labelAlign="left" {...formItemLayout} className={styles.config_form}>
          <Form.Item label="统计间隔">
            <InputNumber
              min={30}
              max={60}
              value={config.refresh_interval || 30}
              onChange={val => handleConfigChange('refresh_interval', val)}
              formatter={val => `${val} min`}
              parser={val => (val || '').replace(' min', '')}
            />
          </Form.Item>
          <Form.Item label="保留时间">
            <InputNumber
              min={30}
              max={60}
              value={config.keep_duration || 30}
              onChange={val => handleConfigChange('keep_duration', val)}
              formatter={val => `${val} day`}
              parser={val => (val || '').replace(' day', '')}
            />
          </Form.Item>
          <Form.Item label="保存 SQL 条数">
            <InputNumber
              min={10}
              max={100}
              value={config.max_sql_count || 100}
              onChange={val => handleConfigChange('max_sql_count', val)}
            />
          </Form.Item>
          <Form.Item label="保留 SQL 最大长度">
            <InputNumber
              min={10}
              max={4096}
              value={config.max_sql_length || 4096}
              onChange={val => handleConfigChange('max_sql_length', val)}
            />
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
}

export default StatementSettingModal;
