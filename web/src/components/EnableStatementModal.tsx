import React, { useState, useEffect } from 'react';
import { Modal, message, Button } from 'antd';
import moment from 'moment';
import { enableInstanceStatement } from '@/services/statement';

interface Props {
  visible: boolean;
  onClose: () => void;
  onData: () => void;
  onSetting: () => void;

  instanceId: string;
}

function EnableStatementModal({ visible, onClose, onData, onSetting, instanceId }: Props) {
  const [curTime, setCurTime] = useState('');

  useEffect(() => {
    setCurTime(moment().format('YYYY-MM-DD HH:mm:ss'));
    const timer = setInterval(() => {
      setCurTime(moment().format('YYYY-MM-DD HH:mm:ss'));
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  async function handleOk() {
    const res = await enableInstanceStatement(instanceId);
    if (res !== undefined) {
      message.error(`${instanceId} 开启 Statement 统计成功`);
      onData();
      onClose();
    }
  }

  return (
    <Modal visible={visible} onCancel={onClose} onOk={handleOk} title="开启 Statement 统计">
      <div>
        开启前请确认设置：
        <Button type="primary" onClick={onSetting}>
          设置
        </Button>
      </div>
      <div>开始统计时间：{curTime}</div>
      <div style={{ color: 'red' }}>
        注：诊断工具开启关闭 Statement 功能，TiDB 配置将随之开启关闭
      </div>
    </Modal>
  );
}

export default EnableStatementModal;
