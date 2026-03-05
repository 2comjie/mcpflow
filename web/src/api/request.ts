import axios from 'axios';
import { message } from 'antd';

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 60000,
});

request.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err.response?.data?.error || err.message;
    message.error(msg);
    return Promise.reject(err);
  }
);

export default request;
