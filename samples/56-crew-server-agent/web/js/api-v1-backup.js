/**
 * API Service Layer
 * Handles all communication with the Nova Crew Server Agent API
 */

// IMPORTANT: Use port 8081 (CORS proxy) to avoid CORS issues
// The proxy adds necessary CORS headers to all API endpoints
// To use: run `go run cors-proxy.go` in web/ directory
// If you modified the Nova SDK to include CORS headers, change to port 8080
const API_BASE_URL = 'http://localhost:8081';

class CrewServerAPI {
    constructor(baseURL = API_BASE_URL) {
        this.baseURL = baseURL;
        this.eventSource = null;
        this.abortController = null;
    }

    /**
     * Send a message and receive streaming response via SSE
     * @param {string} message - User message
     * @param {Function} onChunk - Callback for each chunk (content, isComplete)
     * @param {Function} onNotification - Callback for tool call notifications
     * @param {Function} onError - Error callback
     * @returns {Promise<void>}
     */
    async sendMessage(message, onChunk, onNotification, onError) {
        try {
            // Close existing connection if any
            this.closeStream();

            const response = await fetch(`${this.baseURL}/completion`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    data: {
                        message: message
                    }
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            while (true) {
                const { value, done } = await reader.read();

                if (done) {
                    break;
                }

                // Decode chunk and add to buffer
                buffer += decoder.decode(value, { stream: true });

                // Process complete lines
                const lines = buffer.split('\n');
                buffer = lines.pop() || ''; // Keep incomplete line in buffer

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        try {
                            const data = JSON.parse(line.substring(6));

                            // Handle tool call notifications
                            if (data.kind === 'tool_call') {
                                if (onNotification) {
                                    onNotification(data);
                                }
                            }
                            // Handle message chunks
                            else if (data.message !== undefined) {
                                const isComplete = data.finish_reason === 'stop';
                                onChunk(data.message, isComplete);

                                if (isComplete) {
                                    return;
                                }
                            }
                        } catch (e) {
                            console.error('Failed to parse SSE data:', e, line);
                        }
                    }
                }
            }
        } catch (error) {
            console.error('Stream error:', error);
            if (onError) {
                onError(error);
            }
        }
    }

    /**
     * Stop the current streaming completion
     * @returns {Promise<Object>}
     */
    async stopCompletion() {
        try {
            const response = await fetch(`${this.baseURL}/completion/stop`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            return await response.json();
        } catch (error) {
            console.error('Failed to stop completion:', error);
            throw error;
        }
    }

    /**
     * Reset conversation memory (keeps system instruction)
     * @returns {Promise<Object>}
     */
    async resetMemory() {
        try {
            const response = await fetch(`${this.baseURL}/memory/reset`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            return await response.json();
        } catch (error) {
            console.error('Failed to reset memory:', error);
            throw error;
        }
    }

    /**
     * Get list of all conversation messages
     * @returns {Promise<Array>}
     */
    async getMessages() {
        try {
            const response = await fetch(`${this.baseURL}/memory/messages/list`);
            const data = await response.json();
            return data.messages || [];
        } catch (error) {
            console.error('Failed to get messages:', error);
            throw error;
        }
    }

    /**
     * Get context size
     * @returns {Promise<number>}
     */
    async getContextSize() {
        try {
            const response = await fetch(`${this.baseURL}/memory/messages/context-size`);
            const data = await response.json();
            return data.context_size || 0;
        } catch (error) {
            console.error('Failed to get context size:', error);
            throw error;
        }
    }

    /**
     * Validate a pending operation (human-in-the-loop)
     * @param {string} operationId - Operation ID to validate
     * @returns {Promise<Object>}
     */
    async validateOperation(operationId) {
        try {
            const response = await fetch(`${this.baseURL}/operation/validate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    operation_id: operationId
                })
            });
            return await response.json();
        } catch (error) {
            console.error('Failed to validate operation:', error);
            throw error;
        }
    }

    /**
     * Cancel a pending operation
     * @param {string} operationId - Operation ID to cancel
     * @returns {Promise<Object>}
     */
    async cancelOperation(operationId) {
        try {
            const response = await fetch(`${this.baseURL}/operation/cancel`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    operation_id: operationId
                })
            });
            return await response.json();
        } catch (error) {
            console.error('Failed to cancel operation:', error);
            throw error;
        }
    }

    /**
     * Reset all pending operations
     * @returns {Promise<Object>}
     */
    async resetOperations() {
        try {
            const response = await fetch(`${this.baseURL}/operation/reset`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            return await response.json();
        } catch (error) {
            console.error('Failed to reset operations:', error);
            throw error;
        }
    }

    /**
     * Get model information
     * @returns {Promise<Object>}
     */
    async getModels() {
        try {
            const response = await fetch(`${this.baseURL}/models`);
            return await response.json();
        } catch (error) {
            console.error('Failed to get models:', error);
            throw error;
        }
    }

    /**
     * Check server health
     * @returns {Promise<Object>}
     */
    async checkHealth() {
        try {
            const response = await fetch(`${this.baseURL}/health`);
            return await response.json();
        } catch (error) {
            console.error('Health check failed:', error);
            throw error;
        }
    }

    /**
     * Close any active stream
     */
    closeStream() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }
        if (this.abortController) {
            this.abortController.abort();
            this.abortController = null;
        }
    }
}

// Export for use in other modules
window.CrewServerAPI = CrewServerAPI;
