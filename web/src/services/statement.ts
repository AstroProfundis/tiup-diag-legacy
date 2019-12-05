import request from '@/utils/request';

export interface IStatementConfig {
  refresh_interval: number;
  keep_duration: number;
  max_sql_count: number;
  max_sql_length: number;
}

// //////////////////////////////////////
// 获取/打开/关闭集群 statement 开关
export async function getInstanceStatementStatus(instanceId: string) {
  return request(`/instances/${instanceId}/statement`);
}

export async function enableInstanceStatement(instanceId: string) {
  return request.put(`/instances/${instanceId}/statement`);
}

export async function disableInstanceStatement(instanceId: string) {
  return request.delete(`/instances/${instanceId}/statement`);
}

// //////////////////////////////////////
// 获取/修改集群 statement 设置
export async function getInstanceStatementConfig(instanceId: string) {
  return request(`/instances/${instanceId}/statement/config`);
}

export async function updateInstanceStatementConfig(instanceId: string, config: IStatementConfig) {
  return request.put(`/instances/${instanceId}/statement/config`, { data: config });
}

// //////////////////////////////////////
// 获取集群 schemas
export async function getInstanceSchemas(instanceId: string) {
  return request(`/instances/${instanceId}/statements/schemas`);
}

// //////////////////////////////////////
// 获取 schema 出现的时间段
export async function getSchemasRanges(instanceId: string, schemas: string[]) {
  const params = {
    schemas: schemas === [] ? undefined : schemas.join(','),
  };
  return request(`/instances/${instanceId}/statements/ranges`, { params });
}

// //////////////////////////////////////
// 获取 schema 的 statement 统计结果
export async function getStatements(
  instanceId: string,
  schemas: string[],
  begin: number | string,
  end: number | string | undefined,
) {
  const params = {
    begin,
    end,
    schemas: schemas === [] ? undefined : schemas.join(','),
  };
  return request(`/instances/${instanceId}/statements`, { params });
}

// //////////////////////////////////////
// 获取 statement 详情
export async function getStatementDetail(
  instanceId: string,
  digest: string,
  schemas: string[],
  begin: number,
  end: number | undefined,
) {
  const params = {
    begin,
    end,
    schemas: schemas === [] ? undefined : schemas.join(','),
  };
  return request(`/instances/${instanceId}/statements/digests/${digest}`, { params });
}
