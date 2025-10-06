package telegram

import (
	"log"
	"regexp"

	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallbackHandler определяет функцию-обработчик callback'а
type CallbackHandler func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error

// CallbackRoute представляет маршрут для callback'а
type CallbackRoute struct {
	Pattern *regexp.Regexp
	Handler CallbackHandler
}

// CallbackRouter управляет маршрутизацией callback'ов
type CallbackRouter struct {
	routes []CallbackRoute
}

// NewCallbackRouter создает новый роутер
func NewCallbackRouter() *CallbackRouter {
	return &CallbackRouter{
		routes: make([]CallbackRoute, 0),
	}
}

// Register регистрирует новый маршрут с регулярным выражением
func (r *CallbackRouter) Register(pattern string, handler CallbackHandler) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	r.routes = append(r.routes, CallbackRoute{
		Pattern: regex,
		Handler: handler,
	})

	return nil
}

// RegisterSimple регистрирует простой маршрут для точного совпадения
func (r *CallbackRouter) RegisterSimple(pattern string, handler CallbackHandler) {
	r.routes = append(r.routes, CallbackRoute{
		Pattern: regexp.MustCompile("^" + regexp.QuoteMeta(pattern) + "$"),
		Handler: handler,
	})
}

// RegisterPrefix регистрирует маршрут для префикса с извлечением параметра
func (r *CallbackRouter) RegisterPrefix(prefix string, handler CallbackHandler) {
	pattern := "^" + regexp.QuoteMeta(prefix) + "(.+)$"
	r.routes = append(r.routes, CallbackRoute{
		Pattern: regexp.MustCompile(pattern),
		Handler: handler,
	})
}

// Handle обрабатывает callback, находя подходящий маршрут
func (r *CallbackRouter) Handle(callback *tgbotapi.CallbackQuery, user *models.User) error {
	data := callback.Data

	for _, route := range r.routes {
		matches := route.Pattern.FindStringSubmatch(data)
		if matches != nil {
			// Извлекаем параметры из совпадений
			params := make(map[string]string)

			// Первый элемент - полное совпадение, остальные - группы
			if len(matches) > 1 {
				// Для префиксных маршрутов сохраняем параметр
				params["param"] = matches[1]
			}

			return route.Handler(callback, user, params)
		}
	}

	log.Printf("DEBUG: No route matched for callback data: '%s'", data)
	return nil // Не найден подходящий обработчик
}

// SetupIsolatedRoutes настраивает маршруты для изолированного редактора интересов
func (r *CallbackRouter) SetupIsolatedRoutes(handler *TelegramHandler) error {
	// Простые маршруты для точного совпадения
	r.RegisterSimple("isolated_edit_start", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedEditStart(callback, user)
	})

	r.RegisterSimple("isolated_main_menu", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedMainMenu(callback, user)
	})

	r.RegisterSimple("isolated_edit_categories", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedEditCategories(callback, user)
	})

	r.RegisterSimple("isolated_edit_primary", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedEditPrimary(callback, user)
	})

	r.RegisterSimple("isolated_preview_changes", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedPreviewChanges(callback, user)
	})

	r.RegisterSimple("isolated_save_changes", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedSaveChanges(callback, user)
	})

	r.RegisterSimple("isolated_cancel_edit", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedCancelEdit(callback, user)
	})

	r.RegisterSimple("isolated_undo_last", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedUndoLast(callback, user)
	})

	r.RegisterSimple("isolated_show_stats", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return handler.HandleIsolatedShowStats(callback, user)
	})

	// Префиксные маршруты с параметрами
	r.RegisterPrefix("isolated_edit_category_", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		categoryKey := params["param"]
		return handler.HandleIsolatedEditCategory(callback, user, categoryKey)
	})

	r.RegisterPrefix("isolated_toggle_interest_", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		interestIDStr := params["param"]
		return handler.HandleIsolatedToggleInterest(callback, user, interestIDStr)
	})

	r.RegisterPrefix("isolated_toggle_primary_", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		interestIDStr := params["param"]
		return handler.HandleIsolatedTogglePrimary(callback, user, interestIDStr)
	})

	r.RegisterPrefix("isolated_select_all_", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		categoryKey := params["param"]
		return handler.HandleIsolatedMassSelect(callback, user, categoryKey)
	})

	r.RegisterPrefix("isolated_clear_all_", func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		categoryKey := params["param"]
		return handler.HandleIsolatedMassClear(callback, user, categoryKey)
	})

	return nil
}
