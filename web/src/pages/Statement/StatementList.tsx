import React, { useState, useEffect } from 'react';
import { Select, Button, Modal } from 'antd';
import { connect } from 'dva';
import { ConnectProps, ConnectState, Dispatch } from '@/models/connect';
import { IFormatInstance } from '@/models/inspection';
import { StatementModelState } from '@/models/statement';
import { getInstanceStatementStatus, disableInstanceStatement } from '@/services/statement';
import EnableStatementModal from '@/components/EnableStatementModal';
import StatementSettingModal from '@/components/StatementSettingModal';

const { Option } = Select;

const styles = require('../style.less');

type EnableStatmentStatus = 'on' | 'off' | 'unknown';

interface StatementListProps extends ConnectProps {
  dispatch: Dispatch;

  instances: IFormatInstance[];
  loadingInstances: boolean;

  statement: StatementModelState;
  loadingStatements: boolean;
  loadingSchemas: boolean;
  loadingTimeRanges: boolean;
}

function StatementList({
  dispatch,
  instances,
  loadingInstances,
  statement,
  loadingStatements,
  loadingSchemas,
  loadingTimeRanges,
}: StatementListProps) {
  const [enableStatment, setEnableStatement] = useState<EnableStatmentStatus>('unknown');

  const [enableStatementModalVisible, setEnableStatementModalVisible] = useState(false);
  const [statementSettingModalVisible, setStatementSettingModalVisible] = useState(false);

  useEffect(() => {
    dispatch({
      type: 'inspection/fetchInstances',
    });
  }, []);

  function handleInstanceChange(val: string) {
    dispatch({
      type: 'statement/changeInstance',
      payload: val,
    });
    dispatch({
      type: 'statement/fetchSchemas',
    });

    setEnableStatement('off');
    if (val !== undefined) {
      getInstanceStatementStatus(val).then(res => {
        if (val === statement.curInstance && res !== undefined) {
          setEnableStatement('on');
        }
      });
    }
  }

  function handleSchemaChange(val: any) {
    dispatch({
      type: 'statement/changeSchema',
      payload: val,
    });
    dispatch({
      type: 'statement/fetchTimeRanges',
    });
  }

  function handleTimeRangeChange(val: any) {
    dispatch({
      type: 'statement/changeTimeRange',
      payload: val,
    });
    dispatch({
      type: 'statement/fetchStatementList',
    });
  }

  function toggleStatementSwitch(enable: boolean) {
    if (enable) {
      // 打开
      setEnableStatementModalVisible(true);
    } else {
      // 关闭
      Modal.confirm({
        title: '关闭 Statement 统计',
        content: '确认要关闭统计吗？关闭后不留存 statement 统计信息！',
        okText: '关闭',
        okButtonProps: { type: 'danger' },
        onOk() {
          disableInstanceStatement(statement.curInstance || '').then(res => {
            if (res !== undefined) {
              setEnableStatement('off');
            }
          });
        },
        onCancel() {},
      });
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.list_header}>
        <Select
          value={statement.curInstance}
          loading={loadingInstances}
          allowClear
          placeholder="选择集群实例"
          style={{ width: 200, marginLeft: 12 }}
          onChange={handleInstanceChange}
        >
          {instances.map(item => (
            <Option value={item.uuid} key={item.uuid}>
              {item.name}
            </Option>
          ))}
        </Select>
        <Select
          mode="multiple"
          loading={loadingSchemas}
          allowClear
          placeholder="选择 schema"
          style={{ width: 200, marginLeft: 12 }}
          onChange={handleSchemaChange}
        >
          {statement.schemas.map(item => (
            <Option value={item} key={item}>
              {item}
            </Option>
          ))}
        </Select>
        <Select
          value={statement.curTimeRange}
          loading={loadingTimeRanges}
          allowClear
          placeholder="选择时间"
          style={{ width: 200, marginLeft: 12 }}
          onChange={handleTimeRangeChange}
        >
          {statement.timeRanges.map(item => (
            <Option value={item} key={item}>
              {item}
            </Option>
          ))}
        </Select>
        <div className={styles.space} />
        {enableStatment === 'on' && (
          <div>
            <Button
              type="primary"
              style={{ backgroundColor: 'rgba(0,128,0,1)' }}
              onClick={() => toggleStatementSwitch(false)}
            >
              已开启
            </Button>
            <Button type="primary" onClick={() => setStatementSettingModalVisible(true)}>
              设置
            </Button>
          </div>
        )}
        {enableStatment === 'off' && (
          <Button type="danger" onClick={() => toggleStatementSwitch(true)}>
            已关闭
          </Button>
        )}
      </div>
      {enableStatementModalVisible && (
        <EnableStatementModal
          instanceId={statement.curInstance || ''}
          visible={enableStatementModalVisible}
          onClose={() => setEnableStatementModalVisible(false)}
          onData={() => setEnableStatement('on')}
          onSetting={() => setStatementSettingModalVisible(true)}
        />
      )}
      {statementSettingModalVisible && (
        <StatementSettingModal
          instanceId={statement.curInstance || ''}
          visible={statementSettingModalVisible}
          onClose={() => setStatementSettingModalVisible(false)}
        />
      )}
    </div>
  );
}

export default connect(({ inspection, statement, loading }: ConnectState) => ({
  instances: inspection.instances,
  loadingInstances: loading.effects['inspection/fetchInstances'],

  statement,
  loadingStatements: loading.effects['statement/fetchStatementList'],
  loadingSchemas: loading.effects['statement/fetchSchemas'],
  loadingTimeRanges: loading.effects['statement/fetchTimeRanges'],
}))(StatementList);
