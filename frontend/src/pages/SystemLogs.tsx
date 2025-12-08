import React, { useState, useEffect } from 'react';
import { Download, Filter, Search, AlertCircle, Info, AlertTriangle, ChevronLeft, ChevronRight } from 'lucide-react';
import { getLogs } from '../services/api';
import type { LogEntry } from '../types';

export const SystemLogs: React.FC = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [searchTerm, setSearchTerm] = useState('');
  const pageSize = 20;

  const fetchLogs = async (page: number = currentPage) => {
    setLoading(true);
    const offset = (page - 1) * pageSize;
    const result = await getLogs(pageSize, offset);
    setLogs(result.data);
    setTotal(result.total);
    setLoading(false);
  };

  useEffect(() => {
    fetchLogs(currentPage);
    const interval = setInterval(() => fetchLogs(currentPage), 30000); // 每30秒刷新
    return () => clearInterval(interval);
  }, [currentPage]);

  const getLevelBadge = (level: string) => {
    switch(level) {
        case 'INFO':
            return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-700"><Info size={12}/> 信息</span>;
        case 'WARN':
            return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-orange-50 text-orange-700"><AlertTriangle size={12}/> 警告</span>;
        case 'ERROR':
            return <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-700"><AlertCircle size={12}/> 错误</span>;
        default: return null;
    }
  };

  const totalPages = Math.ceil(total / pageSize);
  const startRecord = total === 0 ? 0 : (currentPage - 1) * pageSize + 1;
  const endRecord = Math.min(currentPage * pageSize, total);

  // 过滤日志
  const filteredLogs = logs.filter(log =>
    searchTerm === '' ||
    log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
    log.device_id.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handlePreviousPage = () => {
    if (currentPage > 1) {
      setCurrentPage(currentPage - 1);
    }
  };

  const handleNextPage = () => {
    if (currentPage < totalPages) {
      setCurrentPage(currentPage + 1);
    }
  };

  const handlePageClick = (page: number) => {
    setCurrentPage(page);
  };

  // 生成页码按钮
  const renderPageNumbers = () => {
    const pages = [];
    const maxVisiblePages = 5;

    if (totalPages <= maxVisiblePages) {
      // 如果总页数少于最大可见页数，显示所有页码
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // 否则智能显示页码
      if (currentPage <= 3) {
        pages.push(1, 2, 3, 4, '...', totalPages);
      } else if (currentPage >= totalPages - 2) {
        pages.push(1, '...', totalPages - 3, totalPages - 2, totalPages - 1, totalPages);
      } else {
        pages.push(1, '...', currentPage - 1, currentPage, currentPage + 1, '...', totalPages);
      }
    }

    return pages.map((page, index) => {
      if (page === '...') {
        return (
          <span key={`ellipsis-${index}`} className="px-3 py-1 text-slate-400">
            ...
          </span>
        );
      }
      return (
        <button
          key={page}
          onClick={() => handlePageClick(page as number)}
          className={`px-3 py-1 rounded-lg text-sm font-medium transition-colors ${
            currentPage === page
              ? 'bg-primary-600 text-white'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          {page}
        </button>
      );
    });
  };

  return (
    <div className="p-4 md:p-8 space-y-6 md:space-y-8 animate-[fadeIn_0.5s_ease-out]">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <header>
            <h1 className="text-xl md:text-2xl font-bold text-slate-800">系统日志</h1>
            <p className="text-sm md:text-base text-slate-500">审计跟踪与设备事件</p>
        </header>
        <div className="flex gap-2 md:gap-3">
            <button className="flex items-center gap-1.5 md:gap-2 px-3 md:px-4 py-2 bg-white border border-slate-200 rounded-xl text-slate-600 text-xs md:text-sm font-medium hover:bg-slate-50 transition-colors shadow-sm">
                <Filter size={14} className="md:w-4 md:h-4" /> <span className="hidden sm:inline">筛选</span>
            </button>
            <button className="flex items-center gap-1.5 md:gap-2 px-3 md:px-4 py-2 bg-white border border-slate-200 rounded-xl text-slate-600 text-xs md:text-sm font-medium hover:bg-slate-50 transition-colors shadow-sm">
                <Download size={14} className="md:w-4 md:h-4" /> <span className="hidden sm:inline">导出 CSV</span>
            </button>
        </div>
      </div>

      <div className="bg-white rounded-2xl md:rounded-3xl shadow-sm border border-slate-100 overflow-hidden">
        {/* Search Bar */}
        <div className="p-4 border-b border-slate-100">
            <div className="relative max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
                <input
                    type="text"
                    placeholder="搜索消息..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 transition-all"
                />
            </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
            <table className="w-full min-w-[640px]">
                <thead className="bg-slate-50/50">
                    <tr>
                        <th className="px-3 md:px-6 py-3 md:py-4 text-left text-[10px] md:text-xs font-bold text-slate-500 uppercase tracking-wider">时间戳</th>
                        <th className="px-3 md:px-6 py-3 md:py-4 text-left text-[10px] md:text-xs font-bold text-slate-500 uppercase tracking-wider">级别</th>
                        <th className="px-3 md:px-6 py-3 md:py-4 text-left text-[10px] md:text-xs font-bold text-slate-500 uppercase tracking-wider">设备 ID</th>
                        <th className="px-3 md:px-6 py-3 md:py-4 text-left text-[10px] md:text-xs font-bold text-slate-500 uppercase tracking-wider">消息内容</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                    {loading && logs.length === 0 ? (
                      <tr>
                        <td colSpan={4} className="px-6 py-12 text-center text-slate-500">
                          <div className="flex flex-col items-center gap-2">
                            <div className="w-8 h-8 border-3 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
                            <span className="text-sm">加载日志中...</span>
                          </div>
                        </td>
                      </tr>
                    ) : filteredLogs.length === 0 ? (
                      <tr>
                        <td colSpan={4} className="px-6 py-12 text-center text-sm text-slate-500">
                          {searchTerm ? '没有匹配的日志记录' : '暂无日志记录'}
                        </td>
                      </tr>
                    ) : (
                      filteredLogs.map((log) => (
                        <tr key={log.id} className="hover:bg-slate-50/50 transition-colors">
                            <td className="px-3 md:px-6 py-3 md:py-4 text-[11px] md:text-sm text-slate-500 font-mono">
                                {new Date(log.timestamp).toLocaleString('zh-CN', {
                                  month: '2-digit',
                                  day: '2-digit',
                                  hour: '2-digit',
                                  minute: '2-digit'
                                })}
                            </td>
                            <td className="px-3 md:px-6 py-3 md:py-4 whitespace-nowrap">
                                {getLevelBadge(log.level)}
                            </td>
                            <td className="px-3 md:px-6 py-3 md:py-4 text-[11px] md:text-sm font-medium text-slate-700">
                                <div className="max-w-[80px] md:max-w-none truncate">{log.device_id}</div>
                            </td>
                            <td className="px-3 md:px-6 py-3 md:py-4 text-[11px] md:text-sm text-slate-600">
                                <div className="max-w-[200px] md:max-w-none line-clamp-2 md:line-clamp-none">{log.message}</div>
                            </td>
                        </tr>
                      ))
                    )}
                </tbody>
            </table>
        </div>

        {/* Pagination Footer */}
        <div className="px-6 py-4 border-t border-slate-100 flex items-center justify-between">
            <span className="text-sm text-slate-500">
              显示 {startRecord}-{endRecord} 条，共 {total} 条结果
            </span>

            {totalPages > 1 && (
              <div className="flex items-center gap-2">
                <button
                  onClick={handlePreviousPage}
                  disabled={currentPage === 1}
                  className={`p-1.5 rounded-lg transition-colors ${
                    currentPage === 1
                      ? 'text-slate-300 cursor-not-allowed'
                      : 'text-slate-600 hover:bg-slate-100'
                  }`}
                >
                  <ChevronLeft size={18} />
                </button>

                <div className="flex items-center gap-1">
                  {renderPageNumbers()}
                </div>

                <button
                  onClick={handleNextPage}
                  disabled={currentPage === totalPages}
                  className={`p-1.5 rounded-lg transition-colors ${
                    currentPage === totalPages
                      ? 'text-slate-300 cursor-not-allowed'
                      : 'text-slate-600 hover:bg-slate-100'
                  }`}
                >
                  <ChevronRight size={18} />
                </button>
              </div>
            )}
        </div>
      </div>
    </div>
  );
};
