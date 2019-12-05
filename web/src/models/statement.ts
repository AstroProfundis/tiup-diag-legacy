import { Effect } from 'dva';
import { Reducer } from 'redux';
import { getStatements, getInstanceSchemas, getSchemasRanges } from '@/services/statement';

export interface StatementModelState {
  curInstance: string | undefined;
  curSchema: string[];
  curTimeRange: string | undefined;

  schemas: string[];
  timeRanges: string[];

  statementList: any[];
}

const initialState: StatementModelState = {
  curInstance: undefined,
  curSchema: [],
  curTimeRange: undefined,

  schemas: [],
  timeRanges: [],

  statementList: [],
};

// /////

export interface StatementModelType {
  namespace: 'statement';
  state: StatementModelState;
  effects: {
    fetchSchemas: Effect;
    fetchTimeRanges: Effect;
    fetchStatementList: Effect;
  };
  reducers: {
    saveSchemas: Reducer<StatementModelState>;
    saveTimeRanges: Reducer<StatementModelState>;
    saveStatementList: Reducer<StatementModelState>;

    changeInstance: Reducer<StatementModelState>;
    changeSchema: Reducer<StatementModelState>;
    changeTimeRange: Reducer<StatementModelState>;
  };
}

// /////

const StatementModel: StatementModelType = {
  namespace: 'statement',

  state: initialState,

  effects: {
    *fetchSchemas({ _ }, { call, put, select }) {
      const { curInstance } = yield select((state: any) => state.statement);
      if (curInstance === undefined) {
        return;
      }
      const res = yield call(getInstanceSchemas, curInstance);
      if (res !== undefined) {
        yield put({
          type: 'saveSchemas',
          payload: res,
        });
      }
    },
    *fetchTimeRanges({ _ }, { call, put, select }) {
      const { curInstance, curSchema } = yield select((state: any) => state.statement);
      if (curInstance === undefined) {
        return;
      }
      const res = yield call(getSchemasRanges, curInstance, curSchema);
      if (res !== undefined) {
        yield put({
          type: 'saveTimeRanges',
          payload: res,
        });
      }
    },
    *fetchStatementList({ payload }, { call, put, select }) {
      const { curInstance, curSchema, curTimeRange } = yield select(
        (state: any) => state.statement,
      );
      if (curInstance === undefined || curTimeRange === undefined) {
        return;
      }
      const [begin, end] = curTimeRange.split('|');
      const res = yield call(getStatements, curInstance, curSchema, begin, end);
      if (res !== undefined) {
        yield put({
          type: 'saveStatementList',
          payload: res,
        });
      }
    },
  },
  reducers: {
    saveSchemas(state = initialState, { payload }) {
      return {
        ...state,
        schemas: payload,
      };
    },
    saveTimeRanges(state = initialState, { payload }) {
      return {
        ...state,
        timeRanges: payload,
      };
    },
    saveStatementList(state = initialState, { payload }) {
      return {
        ...state,
        statementList: payload,
      };
    },

    changeInstance(state = initialState, { payload }) {
      return {
        ...state,
        curInstance: payload,
        curSchema: [],
        curTimeRange: undefined,
        schemas: [],
        timeRanges: [],
      };
    },
    changeSchema(state = initialState, { payload }) {
      return {
        ...state,
        curSchema: payload,
        curTimeRange: undefined,
        timeRanges: [],
      };
    },
    changeTimeRange(state = initialState, { payload }) {
      return {
        ...state,
        curTimeRange: payload,
      };
    },
  },
};

export default StatementModel;
