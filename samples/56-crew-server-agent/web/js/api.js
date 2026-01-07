/**
 * API Service Layer V2
 * Based on SnipWise VSCode extension pattern
 * Improved SSE streaming handling
 */

const API_BASE_URL = 'http://localhost:8080';

class CrewServerAPI {
    constructor(baseURL = API_BASE_URL) {
        this.baseURL = baseURL;
        this.currentReader = null;
        this.abortController = null;
    }

    /**
     * Parse response that may be in SSE format or plain JSON
     *
     * SSE format endpoints (use parseResponse):
     * - /operation/validate
     * - /operation/cancel
     * - /operation/reset
     *
     * Plain JSON endpoints (use response.json()):
     * - /models
     * - /memory/messages/context-size
     * - /memory/messages/list
     * - /memory/reset
     * - /completion/stop
     * - /health
     */
    async parseResponse(response, logPrefix = '') {
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const text = await response.text();
        if (logPrefix) {
            console.log(`${logPrefix} raw response:`, text);
        }

        // Parse SSE format: "data: {...}"
        if (text.startsWith('data: ')) {
            const jsonData = text.substring(6).trim();
            const data = JSON.parse(jsonData);
            if (logPrefix) {
                console.log(`${logPrefix} parsed:`, data);
            }
            return data;
        } else {
            // Plain JSON
            const data = JSON.parse(text);
            if (logPrefix) {
                console.log(`${logPrefix} parsed:`, data);
            }
            return data;
        }
    }

    /**
     * Send a message and receive streaming response via SSE
     * Pattern based on SnipWise extension
     */
    async sendMessage(message, onChunk, onNotification, onError) {
        try {
            // Close existing stream
            this.closeStream();

            // Create abort controller for cancellation
            this.abortController = new AbortController();

            const requestBody = JSON.stringify({
                data: {
                    message: message
                }
            });

            const response = await fetch(`${this.baseURL}/completion`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'text/event-stream'
                },
                body: requestBody,
                signal: this.abortController.signal
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            // Get reader for streaming
            this.currentReader = response.body.getReader();
            const decoder = new TextDecoder('utf-8');

            let buffer = '';
            let hasCompleted = false;

            while (!hasCompleted) {
                const { value, done } = await this.currentReader.read();

                if (done) {
                    console.log('Stream completed (done=true)');
                    // Call onChunk with isComplete=true to unlock UI
                    if (onChunk) {
                        onChunk('', true);
                    }
                    break;
                }

                // Decode chunk and add to buffer
                buffer += decoder.decode(value, { stream: true });

                // Split by newlines and process complete lines
                const lines = buffer.split('\n');

                // Keep the last incomplete line in buffer
                buffer = lines.pop() || '';

                for (const line of lines) {
                    // Skip empty lines
                    if (!line.trim()) {
                        continue;
                    }

                    // Process SSE data lines
                    if (line.startsWith('data: ')) {
                        const jsonData = line.substring(6).trim();

                        // Skip empty data
                        if (!jsonData) {
                            continue;
                        }

                        try {
                            const parsed = JSON.parse(jsonData);
                            // console.log('SSE event received:', parsed); // Commenté pour réduire les logs

                            // Handle tool call notifications
                            if (parsed.kind === 'tool_call') {
                                console.log('🔔 Tool call:', parsed.status, parsed.operation_id);
                                if (onNotification) {
                                    onNotification(parsed);
                                }
                                continue;
                            }

                            // Handle information messages (e.g., compression notifications)
                            if (parsed.role === 'information') {
                                console.log('ℹ️ Information:', parsed.content);
                                if (onNotification) {
                                    onNotification(parsed);
                                }
                                continue;
                            }

                            // Handle message chunks
                            if (parsed.message !== undefined) {
                                const chunk = parsed.message;
                                const finishReason = parsed.finish_reason;

                                // console.log('Message chunk:', {chunk: chunk.substring(0, 50), finishReason}); // Commenté

                                // Check for agent switch message
                                const agentSwitchMatch = chunk.match(/<b>Switched to agent: (\w+)\.<\/b>/);
                                if (agentSwitchMatch && onNotification) {
                                    const agentId = agentSwitchMatch[1];
                                    console.log('🔄 Agent switched to:', agentId);
                                    onNotification({
                                        kind: 'agent_switch',
                                        agent_id: agentId
                                    });
                                }

                                // Call chunk callback
                                if (onChunk) {
                                    onChunk(chunk, finishReason === 'stop');
                                }

                                // Check if we're done
                                if (finishReason === 'stop') {
                                    console.log('Stream finished (stop reason)');
                                    hasCompleted = true;
                                    break;
                                }
                            }

                        } catch (parseError) {
                            console.error('Failed to parse JSON:', jsonData, parseError);
                            // Continue processing other lines
                        }
                    }
                }
            }

            // Process any remaining buffer
            if (buffer.trim() && buffer.startsWith('data: ')) {
                try {
                    const jsonData = buffer.substring(6).trim();
                    const parsed = JSON.parse(jsonData);

                    if (parsed.message !== undefined) {
                        onChunk(parsed.message, parsed.finish_reason === 'stop');
                    }
                } catch (e) {
                    console.error('Failed to parse final buffer:', e);
                }
            }

        } catch (error) {
            // Don't report abort as error
            if (error.name === 'AbortError') {
                console.log('Stream aborted by user');
                return;
            }

            console.error('Stream error:', error);
            if (onError) {
                onError(error);
            }
            throw error;
        } finally {
            this.currentReader = null;
        }
    }

    /**
     * Stop the current streaming completion
     */
    async stopCompletion() {
        try {
            // Abort current stream
            if (this.abortController) {
                this.abortController.abort();
                this.abortController = null;
            }

            // Call API stop endpoint
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
     * Reset conversation memory
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
     */
    async getContextSize() {
        try {
            const response = await fetch(`${this.baseURL}/memory/messages/context-size`);
            const data = await response.json();
            return data.characters_count || 0;
        } catch (error) {
            console.error('Failed to get context size:', error);
            throw error;
        }
    }

    /**
     * Validate a pending operation
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

            return await this.parseResponse(response, 'Validation');
        } catch (error) {
            console.error('Failed to validate operation:', error);
            throw error;
        }
    }

    /**
     * Cancel a pending operation
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

            return await this.parseResponse(response, 'Cancel');
        } catch (error) {
            console.error('Failed to cancel operation:', error);
            throw error;
        }
    }

    /**
     * Reset all pending operations
     */
    async resetOperations() {
        try {
            const response = await fetch(`${this.baseURL}/operation/reset`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            return await this.parseResponse(response, 'Reset operations');
        } catch (error) {
            console.error('Failed to reset operations:', error);
            throw error;
        }
    }

    /**
     * Get model information
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
     * Get current agent information
     */
    async getCurrentAgent() {
        try {
            const response = await fetch(`${this.baseURL}/current-agent`);
            return await response.json();
        } catch (error) {
            console.error('Failed to get current agent:', error);
            throw error;
        }
    }

    /**
     * Close any active stream
     */
    closeStream() {
        if (this.currentReader) {
            try {
                this.currentReader.cancel();
            } catch (e) {
                console.error('Error canceling reader:', e);
            }
            this.currentReader = null;
        }

        if (this.abortController) {
            this.abortController.abort();
            this.abortController = null;
        }
    }
}

// Export for use in other modules
window.CrewServerAPI = CrewServerAPI;
