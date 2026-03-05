import request from './request';

export const statsApi = {
  get: () => request.get('/stats'),
};
