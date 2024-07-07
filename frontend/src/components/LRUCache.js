import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Sun, Moon, Trash2 } from 'lucide-react';

const API_BASE_URL = 'http://localhost:8080';

const Card = ({ children, title }) => (
  <motion.div
    className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 mb-6"
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5 }}
  >
    <h3 className="text-xl font-bold mb-4 text-gray-800 dark:text-white">{title}</h3>
    <div>{children}</div>
  </motion.div>
);

const Input = ({ label, ...props }) => (
  <div className="mb-4">
    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{label}</label>
    <input
      {...props}
      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
    />
  </div>
);

const Button = ({ children, className, ...props }) => (
  <motion.button
    whileHover={{ scale: 1.05 }}
    whileTap={{ scale: 0.95 }}
    className={`px-4 py-2 rounded-md text-white font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 ${className}`}
    {...props}
  >
    {children}
  </motion.button>
);

const LRUCache = () => {
  const [key, setKey] = useState('');
  const [value, setValue] = useState('');
  const [expiration, setExpiration] = useState('');
  const [cacheState, setCacheState] = useState([]);
  const [operationResult, setOperationResult] = useState('');
  const [isDarkMode, setIsDarkMode] = useState(() => {
    if (typeof window !== 'undefined') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return false;
  });
  const [error, setError] = useState('');
  const [wsStatus, setWsStatus] = useState('Connecting...');
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    let ws;
    const connectWebSocket = () => {
      ws = new WebSocket('ws://localhost:8080/ws');
      
      ws.onopen = () => {
        console.log('WebSocket connected');
        setWsStatus('Connected');
        setError('');
      };
      ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        setCacheState(data);
      };
      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setWsStatus('Error');
        setError('Failed to connect to WebSocket server. Real-time updates may not be available.');
      };
      ws.onclose = () => {
        setWsStatus('Disconnected');
        setError('WebSocket connection closed. Attempting to reconnect...');
        setTimeout(connectWebSocket, 5000);
      };
    };

    connectWebSocket();

    return () => {
      if (ws) ws.close();
    };
  }, []);

  useEffect(() => {
    if (typeof document !== 'undefined') {
      document.documentElement.classList.toggle('dark', isDarkMode);
    }
  }, [isDarkMode]);

  const handleOperation = useCallback(async (operation, url, method, body = null) => {
    setIsLoading(true);
    try {
      const response = await fetch(url, {
        method,
        headers: body ? { 'Content-Type': 'application/json' } : {},
        body: body ? JSON.stringify(body) : null
      });
  
      const contentType = response.headers.get("content-type");
      let result;
  
      if (contentType && contentType.indexOf("application/json") !== -1) {
        result = await response.json();
      } else {
        result = await response.text();
      }
  
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}, message: ${result}`);
      }
  
      setOperationResult(`${operation} Operation\n${JSON.stringify(result, null, 2)}`);
      setError('');
    } catch (error) {
      console.error(`Error during ${operation} operation:`, error);
      setOperationResult(`Error: ${error.message}`);
      setError(`Failed to perform ${operation} operation. ${error.message}`);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleGet = useCallback(() => {
    if (!key) {
      setError('Please enter a key to GET');
      return;
    }
    handleOperation('GET', `${API_BASE_URL}/api/cache/${key}`, 'GET');
  }, [key, handleOperation]);

  const handleSet = useCallback(() => {
    if (!key || !value || !expiration) {
      setError('Please fill in all fields to SET');
      return;
    }
    if (parseInt(expiration) <= 0) {
      setError('Expiration time must be a positive number');
      return;
    }
    handleOperation('SET', `${API_BASE_URL}/api/cache`, 'POST', { key, value, expiration: parseInt(expiration) });
    setKey('');
    setValue('');
    setExpiration('');
  }, [key, value, expiration, handleOperation]);

  const handleDelete = useCallback(() => {
    if (!key) {
      setError('Please enter a key to DELETE');
      return;
    }
    handleOperation('DELETE', `${API_BASE_URL}/api/cache/${key}`, 'DELETE');
    setKey('');
  }, [key, handleOperation]);

  const toggleTheme = useCallback(() => {
    setIsDarkMode(prevMode => !prevMode);
  }, []);

  return (
    <div className={`min-h-screen p-8 transition-colors duration-300 ${isDarkMode ? 'dark bg-gray-900' : 'bg-gray-100'}`}>
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-4xl font-bold text-gray-800 dark:text-white">LRU Cache Visualizer</h1>
          <div className="flex items-center space-x-4">
            <span className={`text-sm ${wsStatus === 'Connected' ? 'text-green-500' : 'text-red-500'}`}>
              WebSocket: {wsStatus}
            </span>
            <Button onClick={toggleTheme} className="bg-blue-500 hover:bg-blue-600">
              {isDarkMode ? <Sun size={20} /> : <Moon size={20} />}
            </Button>
          </div>
        </div>

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-6" role="alert">
            <strong className="font-bold">Error: </strong>
            <span className="block sm:inline">{error}</span>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card title="Cache Operations">
            <Input label="Key" value={key} onChange={(e) => setKey(e.target.value)} placeholder="Enter key" />
            <Input label="Value" value={value} onChange={(e) => setValue(e.target.value)} placeholder="Enter value" />
            <Input label="Expiration (seconds)" value={expiration} onChange={(e) => setExpiration(e.target.value)} placeholder="Enter expiration time" type="number" />
            <div className="flex space-x-4">
              <Button onClick={handleGet} className="bg-green-500 hover:bg-green-600">GET</Button>
              <Button onClick={handleSet} className="bg-blue-500 hover:bg-blue-600">SET</Button>
              <Button onClick={handleDelete} className="bg-red-500 hover:bg-red-600">
                <Trash2 size={20} />
              </Button>
            </div>
          </Card>

          <Card title="Cache State">
            <AnimatePresence>
              {cacheState.map((item, index) => {
                const expirationTime = new Date(item.expiration);
                const now = new Date();
                const timeLeft = Math.max(0, Math.floor((expirationTime - now) / 1000));
                const progress = Math.max(0, Math.min(100, (timeLeft / 100) * 100)); // Assuming max expiration is 60 seconds
                
                return (
                  <motion.div
                    key={item.key}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: 20 }}
                    transition={{ duration: 0.3 }}
                    className={`bg-gray-200 dark:bg-gray-700 p-4 rounded-md mb-2 ${
                      index === 0 ? 'border-l-4 border-green-500' : ''
                    }`}
                  >
                    <div className="flex justify-between items-center">
                      <div>
                        <span className="font-medium">{item.key}:</span> {item.value}
                      </div>
                      <div className="text-sm text-gray-600 dark:text-gray-400">
                        Expires in: {timeLeft}s
                      </div>
                    </div>
                    <div className="mt-2 bg-gray-300 dark:bg-gray-600 rounded-full h-2">
                      <div
                        className="bg-blue-500 h-2 rounded-full"
                        style={{ width: `${progress}%` }}
                      ></div>
                    </div>
                  </motion.div>
                );
              })}
            </AnimatePresence>
          </Card>

          <Card title="Operation Result">
            <pre className="bg-gray-100 dark:bg-gray-800 p-4 rounded-md overflow-x-auto whitespace-pre-wrap text-gray-800 dark:text-gray-200">
              {operationResult}
            </pre>
          </Card>

          <Card title="How It Works">
            <p className="text-gray-700 dark:text-gray-300 mb-4">
              The LRU (Least Recently Used) Cache efficiently manages a fixed number of items. When you
              GET an item, it moves to the front of the cache. SET operations add new items to the front or
              update existing ones. If the cache is full, the least recently used item is removed to make space.
              Items also have expiration times, ensuring data freshness.
            </p>
            <p className="text-gray-700 dark:text-gray-300">
              This visualization includes a WebSocket connection, demonstrating real-time updates as you
              interact with the cache. It showcases how the LRU Cache maintains the most relevant data,
              making it invaluable for quick data retrieval in various applications.
            </p>
          </Card>
        </div>
      </div>
      {isLoading && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-blue-500"></div>
        </div>
      )}
    </div>
  );
};

export default LRUCache;