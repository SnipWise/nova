/**
 * InputBar Component
 * User input area with action buttons
 */

const InputBar = {
    name: 'InputBar',

    props: {
        isLoading: {
            type: Boolean,
            default: false
        }
    },

    emits: ['send', 'stop', 'reset-memory', 'show-messages', 'show-models', 'reset-operations', 'copy-last-response'],

    setup(props, { emit }) {
        const userInput = Vue.ref('');

        const handleSend = () => {
            const message = userInput.value.trim();
            if (message && !props.isLoading) {
                emit('send', message);
                userInput.value = '';
            }
        };

        const handleStop = () => {
            emit('stop');
        };

        const handleResetMemory = () => {
            emit('reset-memory');
        };

        const handleShowMessages = () => {
            emit('show-messages');
        };

        const handleShowModels = () => {
            emit('show-models');
        };

        const handleResetOperations = () => {
            emit('reset-operations');
        };

        const handleCopyLastResponse = (event) => {
            emit('copy-last-response', event);
        };

        const handleKeyDown = (event) => {
            // Send on Enter (without Shift)
            if (event.key === 'Enter' && !event.shiftKey) {
                event.preventDefault();
                handleSend();
            }
        };

        const canSend = Vue.computed(() => {
            return userInput.value.trim().length > 0 && !props.isLoading;
        });

        return {
            userInput,
            handleSend,
            handleStop,
            handleResetMemory,
            handleShowMessages,
            handleShowModels,
            handleResetOperations,
            handleCopyLastResponse,
            handleKeyDown,
            canSend
        };
    },

    template: `
        <div class="input-container">
            <div class="input-wrapper">
                <textarea
                    v-model="userInput"
                    @keydown="handleKeyDown"
                    placeholder="Type your message... (Enter to send, Shift+Enter for new line)"
                    class="input-field"
                    :disabled="isLoading"
                ></textarea>
            </div>

            <div class="button-group">
                <button
                    class="primary"
                    @click="handleSend"
                    :disabled="!canSend"
                >
                    📤 Send
                    <span v-if="isLoading" class="loading"></span>
                </button>

                <button
                    class="danger"
                    @click="handleStop"
                    :disabled="!isLoading"
                >
                    ⏹ Stop
                </button>

                <button
                    class="warning"
                    @click="handleResetMemory"
                    :disabled="isLoading"
                >
                    🗑 Clear Memory
                </button>

                <button
                    @click="handleShowMessages"
                    :disabled="isLoading"
                >
                    💬 View Messages
                </button>

                <button
                    @click="handleShowModels"
                    :disabled="isLoading"
                >
                    🤖 View Models
                </button>

                <button
                    class="warning"
                    @click="handleResetOperations"
                    :disabled="isLoading"
                >
                    🔄 Reset Operations
                </button>

                <button
                    class="info"
                    @click="handleCopyLastResponse"
                    :disabled="isLoading"
                    title="Copy last assistant response to clipboard"
                >
                    📋 Copy Response
                </button>
            </div>
        </div>
    `
};

// Register component globally
window.InputBar = InputBar;
