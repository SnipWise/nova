/**
 * Main Application
 * Vue.js 3 App for Nova Crew Server Agent Chat Interface
 */

const { createApp } = Vue;

const App = {
    name: 'App',

    components: {
        ChatMessage,
        StatusBar,
        InputBar,
        OperationControls,
        Modal
    },

    setup() {
        // State
        const messages = Vue.ref([]);
        const contextSize = Vue.ref(0);
        const models = Vue.ref({});
        const selectedAgent = Vue.ref('generic');
        const isLoading = Vue.ref(false);
        const error = Vue.ref(null);
        const pendingOperations = Vue.ref([]);
        const streamingMessageIndex = Vue.ref(-1);

        // Modal state
        const showClearMemoryModal = Vue.ref(false);
        const showModelsModal = Vue.ref(false);
        const showMessagesModal = Vue.ref(false);
        const showResetOperationsModal = Vue.ref(false);
        const allMessages = Vue.ref([]);

        // API instance
        const api = new CrewServerAPI();

        // Refs for auto-scroll
        const chatContainer = Vue.ref(null);

        // Lifecycle
        Vue.onMounted(async () => {
            await loadInitialData();
            startContextSizePolling();
        });

        Vue.onBeforeUnmount(() => {
            stopContextSizePolling();
        });

        // Auto-scroll to bottom when new messages arrive
        Vue.watch(() => messages.value.length, () => {
            Vue.nextTick(() => {
                scrollToBottom();
            });
        });

        // Watch for changes in message content (streaming)
        Vue.watch(
            () => messages.value.map(m => m.content).join(''),
            () => {
                Vue.nextTick(() => {
                    scrollToBottom();
                });
            }
        );

        // Methods
        const loadInitialData = async () => {
            try {
                // Load models
                const modelsData = await api.getModels();
                models.value = modelsData;

                // Load current agent
                const currentAgent = await api.getCurrentAgent();
                selectedAgent.value = currentAgent.agent_id;
                // Update models.chat_model with the current agent's model
                if (currentAgent.model_id) {
                    models.value.chat_model = currentAgent.model_id;
                }

                // Load context size
                const size = await api.getContextSize();
                contextSize.value = size;

                // Check health
                await api.checkHealth();
            } catch (err) {
                console.error('Failed to load initial data:', err);
                error.value = 'Failed to connect to server. Please ensure the server is running.';
            }
        };

        let contextSizeInterval = null;

        const startContextSizePolling = () => {
            // Update context size and current agent every 2 seconds
            contextSizeInterval = setInterval(async () => {
                try {
                    const size = await api.getContextSize();
                    contextSize.value = size;

                    // Update current agent info
                    const currentAgent = await api.getCurrentAgent();
                    selectedAgent.value = currentAgent.agent_id;
                    if (currentAgent.model_id) {
                        models.value.chat_model = currentAgent.model_id;
                    }
                } catch (err) {
                    console.error('Failed to update context size:', err);
                }
            }, 2000);
        };

        const stopContextSizePolling = () => {
            if (contextSizeInterval) {
                clearInterval(contextSizeInterval);
                contextSizeInterval = null;
            }
        };

        const scrollToBottom = () => {
            if (chatContainer.value) {
                chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
            }
        };

        const handleSendMessage = async (message) => {
            error.value = null;

            // Add user message to UI
            messages.value.push({
                role: 'user',
                content: message
            });

            // Add empty assistant message for streaming
            const assistantMessageIndex = messages.value.length;
            messages.value.push({
                role: 'assistant',
                content: ''
            });

            streamingMessageIndex.value = assistantMessageIndex;
            isLoading.value = true;

            try {
                await api.sendMessage(
                    message,
                    // onChunk
                    (chunk, isComplete) => {
                        if (messages.value[assistantMessageIndex]) {
                            messages.value[assistantMessageIndex].content += chunk;
                        }

                        if (isComplete) {
                            isLoading.value = false;
                            streamingMessageIndex.value = -1;
                        }
                    },
                    // onNotification
                    (notification) => {
                        handleToolCallNotification(notification);
                    },
                    // onError
                    (err) => {
                        error.value = err.message || 'An error occurred during streaming';
                        isLoading.value = false;
                        streamingMessageIndex.value = -1;
                    }
                );
            } catch (err) {
                error.value = err.message || 'Failed to send message';
                isLoading.value = false;
                streamingMessageIndex.value = -1;
            }
        };

        const handleStopCompletion = async () => {
            try {
                await api.stopCompletion();
                isLoading.value = false;
                streamingMessageIndex.value = -1;
            } catch (err) {
                error.value = 'Failed to stop completion';
            }
        };

        const handleResetMemory = () => {
            showClearMemoryModal.value = true;
        };

        const confirmResetMemory = async () => {
            try {
                await api.resetMemory();
                messages.value = [];
                contextSize.value = 0;
                pendingOperations.value = [];
                error.value = null;
                showClearMemoryModal.value = false;
            } catch (err) {
                error.value = 'Failed to reset memory';
            }
        };

        const handleShowMessages = async () => {
            try {
                const msgs = await api.getMessages();
                console.log('Messages received:', msgs);
                allMessages.value = msgs;
                showMessagesModal.value = true;
            } catch (err) {
                console.error('Error fetching messages:', err);
                error.value = 'Failed to fetch messages';
            }
        };

        const handleShowModels = async () => {
            try {
                const modelsData = await api.getModels();
                models.value = modelsData;
                showModelsModal.value = true;
            } catch (err) {
                error.value = 'Failed to fetch models';
            }
        };

        const handleResetOperations = () => {
            showResetOperationsModal.value = true;
        };

        const confirmResetOperations = async () => {
            try {
                await api.resetOperations();
                pendingOperations.value = [];
                showResetOperationsModal.value = false;
            } catch (err) {
                error.value = 'Failed to reset operations';
            }
        };

        const handleToolCallNotification = (notification) => {
            console.log('Notification received:', notification);

            // Handle agent switch notifications
            if (notification.kind === 'agent_switch') {
                selectedAgent.value = notification.agent_id;
                console.log('Agent switched to:', notification.agent_id);
                return;
            }

            // Handle tool call notifications
            if (notification.kind === 'tool_call') {
                // Find existing operation or create new one
                const existingIndex = pendingOperations.value.findIndex(
                    op => op.operation_id === notification.operation_id
                );

                if (existingIndex !== -1) {
                    // Update existing operation
                    pendingOperations.value[existingIndex] = notification;
                } else {
                    // Add new operation
                    pendingOperations.value.push(notification);
                }

                // Remove completed/cancelled operations after a delay
                if (notification.status === 'completed' || notification.status === 'cancelled') {
                    setTimeout(() => {
                        const index = pendingOperations.value.findIndex(
                            op => op.operation_id === notification.operation_id
                        );
                        if (index !== -1) {
                            pendingOperations.value.splice(index, 1);
                        }
                    }, 3000);
                }
            }
        };

        const handleValidateOperation = async (operationId) => {
            try {
                console.log('Validating operation:', operationId);
                const result = await api.validateOperation(operationId);
                console.log('Validation result:', result);

                // Update operation status
                const operation = pendingOperations.value.find(
                    op => op.operation_id === operationId
                );
                if (operation) {
                    operation.status = 'completed';
                    operation.message = 'Operation validated';

                    // Remove after delay
                    setTimeout(() => {
                        const index = pendingOperations.value.findIndex(
                            op => op.operation_id === operationId
                        );
                        if (index !== -1) {
                            pendingOperations.value.splice(index, 1);
                        }
                    }, 3000);
                }
            } catch (err) {
                console.error('Validation error:', err);
                error.value = `Failed to validate operation: ${err.message}`;
            }
        };

        const handleCancelOperation = async (operationId) => {
            try {
                console.log('Cancelling operation:', operationId);
                const result = await api.cancelOperation(operationId);
                console.log('Cancel result:', result);

                // Update operation status
                const operation = pendingOperations.value.find(
                    op => op.operation_id === operationId
                );
                if (operation) {
                    operation.status = 'cancelled';
                    operation.message = 'Operation cancelled';

                    // Remove after delay
                    setTimeout(() => {
                        const index = pendingOperations.value.findIndex(
                            op => op.operation_id === operationId
                        );
                        if (index !== -1) {
                            pendingOperations.value.splice(index, 1);
                        }
                    }, 3000);
                }
            } catch (err) {
                console.error('Cancel error:', err);
                error.value = `Failed to cancel operation: ${err.message}`;
            }
        };

        const isMessageStreaming = (index) => {
            return streamingMessageIndex.value === index;
        };

        const hasMessages = Vue.computed(() => messages.value.length > 0);

        return {
            messages,
            contextSize,
            models,
            selectedAgent,
            isLoading,
            error,
            pendingOperations,
            chatContainer,
            handleSendMessage,
            handleStopCompletion,
            handleResetMemory,
            confirmResetMemory,
            handleShowMessages,
            handleShowModels,
            handleResetOperations,
            confirmResetOperations,
            handleValidateOperation,
            handleCancelOperation,
            isMessageStreaming,
            hasMessages,
            // Modal state
            showClearMemoryModal,
            showModelsModal,
            showMessagesModal,
            showResetOperationsModal,
            allMessages
        };
    },

    template: `
        <div>
            <header class="header">
                <h1>🚀 Nova Crew Server Agent</h1>
                <status-bar
                    :context-size="contextSize"
                    :models="models"
                    :selected-agent="selectedAgent"
                />
            </header>

            <div class="chat-container" ref="chatContainer">
                <div v-if="error" class="error">
                    ⚠️ {{ error }}
                </div>

                <div v-if="!hasMessages" class="empty-state">
                    <h2>👋 Welcome!</h2>
                    <p>Start a conversation by typing a message below.</p>
                </div>

                <chat-message
                    v-for="(message, index) in messages"
                    :key="index"
                    :message="message"
                    :is-streaming="isMessageStreaming(index)"
                />
            </div>

            <div class="operations-overlay">
                <operation-controls
                    v-for="operation in pendingOperations"
                    :key="operation.operation_id"
                    :operation="operation"
                    @validate="handleValidateOperation"
                    @cancel="handleCancelOperation"
                />
            </div>

            <input-bar
                :is-loading="isLoading"
                @send="handleSendMessage"
                @stop="handleStopCompletion"
                @reset-memory="handleResetMemory"
                @show-messages="handleShowMessages"
                @show-models="handleShowModels"
                @reset-operations="handleResetOperations"
            />

            <!-- Clear Memory Confirmation Modal -->
            <modal
                :show="showClearMemoryModal"
                title="Clear Memory"
                type="confirm"
                @close="showClearMemoryModal = false"
                @confirm="confirmResetMemory"
                @cancel="showClearMemoryModal = false"
            >
                <p>Are you sure you want to clear the conversation memory?</p>
                <p style="color: #9e9e9e; font-size: 0.875rem; margin-top: 0.5rem;">
                    This will delete all messages and reset the context. This action cannot be undone.
                </p>
            </modal>

            <!-- Models Info Modal -->
            <modal
                :show="showModelsModal"
                title="Models Configuration"
                type="info"
                @close="showModelsModal = false"
            >
                <div class="modal-list-item" v-for="(value, key) in models" :key="key">
                    <div class="modal-list-label">{{ key }}</div>
                    <div class="modal-list-value">{{ value }}</div>
                </div>
            </modal>

            <!-- Messages List Modal -->
            <modal
                :show="showMessagesModal"
                title="Conversation Messages"
                type="info"
                width="800px"
                @close="showMessagesModal = false"
            >
                <p v-if="allMessages.length === 0" style="color: #9e9e9e; text-align: center; padding: 2rem;">
                    No messages in conversation.
                </p>
                <div v-else>
                    <p style="color: #9e9e9e; font-size: 0.875rem; margin-bottom: 1rem;">
                        Total messages: {{ allMessages.length }}
                    </p>
                    <div class="modal-list-item" v-for="(msg, index) in allMessages" :key="index">
                        <div class="modal-list-label" v-if="msg.role">
                            {{ msg.role }}
                            <span v-if="msg.name" style="color: #4fc3f7; margin-left: 0.5rem;">({{ msg.name }})</span>
                        </div>
                        <div class="modal-list-value" style="white-space: pre-wrap; font-family: monospace; font-size: 0.85rem;">
                            {{ msg.content || msg.message || JSON.stringify(msg, null, 2) }}
                        </div>
                    </div>
                </div>
            </modal>

            <!-- Reset Operations Confirmation Modal -->
            <modal
                :show="showResetOperationsModal"
                title="Reset Operations"
                type="confirm"
                @close="showResetOperationsModal = false"
                @confirm="confirmResetOperations"
                @cancel="showResetOperationsModal = false"
            >
                <p>Are you sure you want to reset all pending operations?</p>
                <p style="color: #9e9e9e; font-size: 0.875rem; margin-top: 0.5rem;">
                    This will clear all tool validation requests. This action cannot be undone.
                </p>
            </modal>
        </div>
    `
};

// Create and mount the app
createApp(App).mount('#app');
