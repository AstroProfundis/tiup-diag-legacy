const Mock = require('mockjs');

const mockedLog = {
  time: '@datetime',
  // filename: '@word',
  // file() {
  //   return `${this.filename}.log`;
  // },
  instance_name: '@name',
  'level|1': ['DEBUG', 'INFO', 'WARNING', 'ERROR', 'OTHERS'],
  content: '@sentence',
};

const getLogInstances = (req: any, res: any) => {
  setTimeout(() => {
    res.send(
      Mock.mock({
        'data|5-10': [
          {
            uuid: '@guid',
            name: '@name',
          },
        ],
      }).data,
    );
  }, 1000);
};

const getLogs = (req: any, res: any) => {
  setTimeout(() => {
    res.send(
      Mock.mock({
        token: '@guid',
        'logs|15': [mockedLog],
      }),
    );
  }, 1000);
};

const uploadLogs = (req: any, res: any) => {
  setTimeout(() => {
    res.send(
      Mock.mock({
        logId: '@guid',
      }),
    );
  }, 1000);
};

export default {
  'GET /api/v1/loginstances': getLogInstances,
  'GET /api/v1/loginstances/:id/logs': getLogs,

  'POST /api/v1/logs': uploadLogs,
  'GET /api/v1/logs/:id': getLogs,
};
