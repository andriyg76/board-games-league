import { useMessage } from 'naive-ui';
import { useI18n } from 'vue-i18n';

/**
 * Composable for handling and displaying errors to users
 * Uses Naive UI message API to show user-friendly error notifications
 */
export function useErrorHandler() {
  const message = useMessage();
  const { t } = useI18n();

  /**
   * Extract error message from various error types
   */
  const getErrorMessage = (error: unknown): string => {
    if (error instanceof Error) {
      // Check if it's an API error with status code
      if (error.message.includes('API request failed:')) {
        const match = error.message.match(/API request failed: (\d+)/);
        if (match) {
          const statusCode = parseInt(match[1]);
          switch (statusCode) {
            case 400:
              return t('errors.badRequest');
            case 401:
              return t('errors.unauthorized');
            case 403:
              return t('errors.forbidden');
            case 404:
              return t('errors.notFound');
            case 409:
              return t('errors.conflict');
            case 500:
              return t('errors.serverError');
            default:
              return t('errors.networkError');
          }
        }
      }
      return error.message;
    }

    if (typeof error === 'string') {
      return error;
    }

    return t('errors.unknown');
  };

  /**
   * Show error message to user and log to console
   */
  const handleError = (error: unknown, customMessage?: string) => {
    // Log full error details to console for debugging
    if (customMessage) {
      console.error(`${customMessage}:`, error);
    } else {
      console.error('Error occurred:', error);
    }

    const errorMessage = customMessage || getErrorMessage(error);
    message.error(errorMessage, {
      duration: 5000,
      keepAliveOnHover: true,
    });
  };

  /**
   * Show success message to user
   */
  const showSuccess = (msg: string) => {
    message.success(msg, {
      duration: 3000,
    });
  };

  /**
   * Show warning message to user
   */
  const showWarning = (msg: string) => {
    message.warning(msg, {
      duration: 4000,
    });
  };

  /**
   * Show info message to user
   */
  const showInfo = (msg: string) => {
    message.info(msg, {
      duration: 3000,
    });
  };

  return {
    handleError,
    showSuccess,
    showWarning,
    showInfo,
  };
}
