package main

import (
	"os"
	"log"
  "fmt"
  "strings"
  "strconv"

	tb "gopkg.in/tucnak/telebot.v2"

  // "database/sql"
  _ "github.com/lib/pq"
  "github.com/jmoiron/sqlx"
  "github.com/go-redis/redis/v7"
)

var schema = `
  CREATE TABLE IF NOT EXISTS SEworker(
    id SERIAL PRIMARY KEY,
    tid TEXT,
    approved BOOLEAN
  );

  CREATE TABLE IF NOT EXISTS SEproject(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    difficulty INT,
    price INT,
    paid INT,
    progress INT,
    worker_id INT
  );
`

type SEworker struct {
  Id int `db:"id"`
  Tid string `db:"tid"`
  Approved bool `db:"approved"`
}

type SEproject struct {
  Id int `db:"id"`
  Name string `db:"name"`
  Description string `db:"description"`
  Difficulty int `db:"difficulty"`
  Price int `db:"price"`
  Paid int `db:"paid"`
  Progress int `db:"progress"`
  WorkerId int `db:"worker_id"`
}

func parsePsqlElements(url string) (string, string, string, string, string) {
  split := strings.Split(url, "@")
  unamepwdsplit := strings.Split(split[0], "//")
  unamepwd := strings.Split(unamepwdsplit[1], ":")
  uname := unamepwd[0]
  pwd := unamepwd[1]
  urlportdbname := strings.Split(split[1], ":")
  link := urlportdbname[0]
  portdbname := strings.Split(urlportdbname[1], "/")
  port := portdbname[0]
  dbname := portdbname[1]
  return uname, pwd, link, port, dbname
}
// redis://h:pce2cf2e8633a6107d63b9e1aed57cd5a6590af92578cada0f451abc279b13bf7@ec2-18-203-184-0.eu-west-1.compute.amazonaws.com:6549
func parseRedisElements(url string) (string, string, string) {
  split := strings.Split(url, "@")
  unamepwdsplit := strings.Split(split[0], "//")
  unamepwd := strings.Split(unamepwdsplit[1], ":")
  uname := unamepwd[0]
  pwd := unamepwd[1]
  link := split[1]
  return uname, pwd, link
}

func main() {
	var (
		port      = os.Getenv("PORT")       // sets automatically
		publicURL = os.Getenv("PUBLIC_URL") // you must add it to your config vars
		token     = os.Getenv("TOKEN")      // you must add it to your config vars
    redisURL  = os.Getenv("REDIS_URL")
    psqlURL   = os.Getenv("DATABASE_URL")
    dbuname, dbpwd, dblink, dbport, dbname = parsePsqlElements(psqlURL)
    psqlInfo  = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s" +
    " sslmode=disable", dblink, dbport, dbuname, dbpwd, dbname)
    _, repwd, relink = parseRedisElements(redisURL)
    // Cuz I'm too lazy to do it the right way
    projectname  = ""
    projectdesc  = ""
    projectdiff  = 0
    projectprice = 0
    takeProjectStr = 1
	)

  db, err := sqlx.Connect("postgres", psqlInfo)
  if err != nil {
    log.Panic(err)
  }
  defer db.Close()
  db.MustExec(schema)

  client := redis.NewClient(&redis.Options{
		Addr:     relink,
		Password: repwd, // no password set
		DB:       0,  // use default DB
	})
  _, err = client.Ping().Result()
  if err != nil {
    log.Panic(err)
  }

	webhook := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	pref := tb.Settings{
		Token:  token,
		Poller: webhook,
	}

  // here are buttons defined

  backBtn := tb.InlineButton{
    Unique: "back",
    Text:   "↩️ Назад"}

  enterBtn := tb.InlineButton{
    Unique: "enter",
    Text:   "🔑 Войти на биржу"}

  qualifyBtn := tb.InlineButton{
    Unique: "qualify",
    Text:   "🧧 Подать заявку"}

  infoBtn := tb.InlineButton{
    Unique: "info",
    Text:   "📃 Информация о бирже"}

  howToEnterBtn := tb.InlineButton{
    Unique: "howToEnter",
    Text:   "🗝 Как попасть на биржу?"}

  fuckedUpBtn := tb.InlineButton{
    Unique: "fuckedUp",
    Text:   "📆 Что будет, если я не уложусь в срок?"}

  whatProjectsBtn := tb.InlineButton{
    Unique: "whatProjects",
    Text:   "📑 Какие проекты предоставляет биржа?"}

  // currentProjectBtn := tb.InlineButton{
  //   Unique: "currentProject",
  //   Text:   "🛎 Мой текущий проект"}

  showOffersBtn := tb.InlineButton{
    Unique: "showOffers",
    Text:   "📜 Посмотреть текущие предложения"}

  askAdminBtn := tb.InlineButton{
    Unique: "askAdmin",
    Text:   "💡 Вопрос администрации"}

  techSuppBtn := tb.InlineButton{
    Unique: "techSupp",
    Text:   "📦 Получить техническую помощь"}

  redeemMilestoneProjectBtn := tb.InlineButton{
    Unique: "redeemMilestoneProject",
    Text:   "✅ Закрыть этап/проект"}

  cancelProjectBtn := tb.InlineButton{
    Unique: "cancelProject",
    Text:   "❌ Отказаться от проекта"}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

  b.Handle("/whoami", func(m *tb.Message) {
    err := client.Set(fmt.Sprintf("%d", m.Sender.ID), "whoami", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(m.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }
    b.Send(m.Sender, fmt.Sprintf("%d", m.Sender.ID))
  })

  b.Handle("/start", func(m *tb.Message) {
    err := client.Set(fmt.Sprintf("%d", m.Sender.ID), "start", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(m.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }
    inlineKeys := [][]tb.InlineButton{
      []tb.InlineButton{enterBtn, qualifyBtn},
      []tb.InlineButton{infoBtn}}

    b.Send(
      m.Sender,
      "Добро пожаловать в Swift Exchange! Пожалуйста, выберите следующий шаг:",
      &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
  })

  b.Handle(&infoBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "info", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    b.Respond(c, &tb.CallbackResponse{
      ShowAlert: false,
    })

    inlineKeys := [][]tb.InlineButton{
      []tb.InlineButton{howToEnterBtn},
      []tb.InlineButton{fuckedUpBtn,whatProjectsBtn},
      []tb.InlineButton{backBtn}}

    b.Send(c.Sender, `📃 Информация о бирже:

Swift Exchange - приватная биржа для доверенных разработчиков.

Мы берем на себя:

📩 Полное общение с заказчиками
💌 Профессиональную и техническую помощь в любом вопросе
📅 Поднятие вашего рейтинга, как разработчика, развитие личного бренда
📈 Постоянный поток проектов

Биржа забирает 5% с каждого проекта и выплачивает разработчику заработанные деньги сразу после принятия работ заказчиком. Способ оплаты обсуждается с каждым разработчиком отдельно.`,
    &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
  })

  b.Handle(&enterBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "enter", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    inlineKeysOff := [][]tb.InlineButton{[]tb.InlineButton{showOffersBtn}}
    inlineKeysOn := [][]tb.InlineButton{[]tb.InlineButton{redeemMilestoneProjectBtn}, []tb.InlineButton{cancelProjectBtn}}

    user := SEworker{}
    err = db.Get(&user, "SELECT * FROM SEworker WHERE tid=$1", c.Sender.ID)
    if err != nil || user.Approved != true {
      b.Send(c.Sender, `Сначала надо пройти собеседование, для этого нажми на "🧧 Подать заявку"`)
      return
    }
    projects := []SEproject{}
    cproject := SEproject{}
    db.Select(&projects, "SELECT * FROM SEproject WHERE worker_id = 0 ORDER BY id DESC")
    db.Get(&cproject, "SELECT * FROM SEproject WHERE worker_id = $1 ORDER BY id DESC", user.Id)
    if cproject.WorkerId != user.Id {
      b.Send(c.Sender, fmt.Sprintf(`🔑 Войти на биржу:

Вы вошли на Swift Exchange. У вас сейчас %d открытых предложений по проектам.`, len(projects)),
      &tb.ReplyMarkup{InlineKeyboard: inlineKeysOff})
      return
    }
    b.Send(c.Sender, fmt.Sprintf(`🔑 Войти на биржу:

Вы вошли на Swift Exchange. В данный момент вы уже выполняете проект "%s".`, cproject.Name),
      &tb.ReplyMarkup{InlineKeyboard: inlineKeysOn})
  })

  b.Handle(&howToEnterBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "howToEnter", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    inlineKeys := [][]tb.InlineButton{[]tb.InlineButton{backBtn}}

    b.Send(
      c.Sender,
      `🗝 Как попасть на биржу?:

Любой разработчик может попасть на биржу. Для этого достаточно нажать на соответствующую кнопку в начале диалога, рассказать пару слов о себе и с Вами свяжется администрация биржи для небольшого текстового интервью. Если вы уже работали на фрилансе, выполняли проекты и вы добросовестный разработчик, Вы обязательно попадете на биржу. Так же Вам нужно будет оплатить символический вступительный взнос, дабы убедиться в Ваших намерениях в размере 350р.`,
      &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
  })

  b.Handle(&whatProjectsBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "whatProjects", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    inlineKeys := [][]tb.InlineButton{[]tb.InlineButton{backBtn}}

    b.Send(
      c.Sender,
      `📑 Какие проекты предоставляет биржа?:

Биржа предлагает любые проекты, связанные с языком Swift. В основном это iOS/macOS приложения. Так же будут задействованы и другие платформы.

Так же проект не будет выставлен на общее обозрение. Мы обращаемся к каждому разработчику по его внутреннему рейтингу, начиная с высокого. Если разработчику подходит проект - мы его передаем. Если нет - то обсуждаем данный проект уже со следующим разработчиком, пока проект не найдет своего исполнителя.`,
      &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
  })

  b.Handle(&fuckedUpBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "fuckedUp", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    inlineKeys := [][]tb.InlineButton{[]tb.InlineButton{backBtn}}

    b.Send(
      c.Sender,
      `📆 Что будет, если я не уложусь в срок?:

Если Вы понимаете, что опаздываете со сдачей проекта на пару дней - ничего страшного, за это не будет никакого штрафа. Если же Вы понимаете, что опаздываете на дней 5 или больше, вы должны предупредить администрацию об этом за 1 неделю до сдачи проекта. Если состояние кода удовлетворительное, то Вам так же ничего не грозит. Просто снятие с проекта и назначение нового.

В иных ситуациях, администрация будет вынуждена попросить Вас выплатить штраф или заблокировать на бирже.`,
      &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
  })

  b.Handle(&qualifyBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "qualify0", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    b.Send(
      c.Sender,
      `🧧 Подать заявку:

Мы очень рады, что Вы решили попробовать себя в нашей бирже. Пожалуйста, напишите кратко о себе, своем опыте в разработке и проектах, с которыми Вы сталкивались. После этого, в самые кратчайшие сроки с Вами свяжется администрация для интервью в текстовом виде. Будем ждать! 😉`)
  })

  b.Handle(&askAdminBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "askAdmin0", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    b.Send(
      c.Sender,
      `💡 Вопрос администрации:

Введите Ваш запрос и он будет направлен администрации. Пожалуйста, старайтесь детальнее описать Вашу проблему. Запросы в формате "У меня проблема, помогите." рассматриваться не будут.`)
  })

  b.Handle(&techSuppBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "techSupp0", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    b.Send(
      c.Sender,
      `📦 Получить техническую помощь:

Вы должны описать проблему в проекте, с которой столкнулись. Чем больше информации Вы предоставите, тем быстрее получите ответ на Ваш вопрос. Мы стараемся помочь Вам в самый краткий срок.

Формат обращения:

1) Название проекта
2) Описание проблемы
3) Приложенные части кода, залитые на pastebin.com или Github Gist, скриншоты или видео, на которых видно проблему`)
  })

  b.Handle(&redeemMilestoneProjectBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "redeemMilestoneProject0", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }

    b.Send(
      c.Sender,
      `✅ Закрыть этап/проект:

Вы собираетесь закрыть проект или этап. Пожалуйста, заполните форму для закрытия:

1) Название проекта
2) Номер этапа, если закрываете этап
3) hash-номер коммита, который можно запускать для тестирования`)
  })

  b.Handle(&backBtn, func(c *tb.Callback) {
    v, err := client.Get(fmt.Sprintf("%d", c.Sender.ID)).Result()
    if err != nil {
      log.Print(err)
    }

    position := fmt.Sprintf("%s", v)
    switch position {
      case "info":
        err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "start", 0).Err()
        if err != nil {
          log.Print(err)
          b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
          return
        }
        inlineKeys := [][]tb.InlineButton{
          []tb.InlineButton{enterBtn, qualifyBtn},
          []tb.InlineButton{infoBtn}}
        b.Send(
          c.Sender,
          "Добро пожаловать в Swift Exchange! Пожалуйста, выберите следующий шаг:",
          &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
        return
      case "fuckedUp", "whatProjects", "howToEnter":
        err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "info", 0).Err()
        if err != nil {
          log.Print(err)
          b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
          return
        }
        inlineKeys := [][]tb.InlineButton{
          []tb.InlineButton{howToEnterBtn},
          []tb.InlineButton{fuckedUpBtn,whatProjectsBtn},
          []tb.InlineButton{backBtn}}

        b.Send(c.Sender, `📃 Информация о бирже:

    Swift Exchange - приватная биржа для доверенных разработчиков.

    Мы берем на себя:

    📩 Полное общение с заказчиками
    💌 Профессиональную и техническую помощь в любом вопросе
    📅 Поднятие вашего рейтинга, как разработчика, развитие личного бренда
    📈 Постоянный поток проектов

    Биржа забирает 5% от суммы с каждого проекта и выплачивает разработчику заработанные деньги сразу после принятия работ заказчиком. Способ оплаты обсуждается с каждым разработчиком отдельно.`,
        &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
        return
      default:
        err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "start", 0).Err()
        if err != nil {
          log.Print(err)
          b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
          return
        }
        inlineKeys := [][]tb.InlineButton{
          []tb.InlineButton{enterBtn, qualifyBtn},
          []tb.InlineButton{infoBtn}}
        b.Send(
          c.Sender,
          "Добро пожаловать в Swift Exchange! Пожалуйста, выберите следующий шаг:",
          &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
        return
    }
  })

  b.Handle(&showOffersBtn, func(c *tb.Callback) {
    projects := []SEproject{}
    db.Select(&projects, "SELECT * FROM SEproject WHERE worker_id = 0 ORDER BY id DESC")
    if len(projects) < 1 {
      b.Send(c.Sender, "К сожалению, в данный момент на бирже нет проектов. Проверяйте биржу каждый день и, возможно, следующий проект будет ваш!")
      return
    }
    for _, project := range projects {
      inlineKeys := [][]tb.InlineButton{
        []tb.InlineButton{tb.InlineButton{
          Unique: fmt.Sprintf("%d:%d:%d", takeProjectStr, takeProjectStr, project.Id),
          Text:   "❇️ Принять проект"}}}
      b.Send(
        c.Sender,
        fmt.Sprintf(`– %s
Задача: %s
Сложность: %d/5 | Стоимость: %d руб.
`, project.Name, project.Description, project.Difficulty, project.Price),
        &tb.ReplyMarkup{InlineKeyboard: inlineKeys})
    }
    return
  })

  b.Handle(tb.OnCallback, func(c *tb.Callback) {
    split := strings.Split(c.Data, ":")
    pid := split[2]
    cmd, err := strconv.Atoi(split[1])
    if err != nil || cmd != takeProjectStr {
      b.Send(
        c.Sender,
        "https://www.youtube.com/watch?v=l60MnDJklnM")
      return
    }
    project := SEproject{}
    db.Select(&project, `SELECT * FROM SEproject WHERE id = $1`, pid)
    if project.WorkerId != 0 {
      b.Send(
        c.Sender,
        "К сожалению, данный проект уже занят, выберите другой")
      return
    }
    worker := SEworker{}
    db.Get(&worker, `SELECT * FROM SEworker WHERE tid = $1`, c.Sender.ID)
    fmt.Println(worker)
    db.MustExec(`UPDATE SEproject SET worker_id = $1 WHERE id = $2`, worker.Id, pid)
    b.Send(
      c.Sender,
      "Этот проект ваш, скоро с вами свяжется администратор для уточнения деталей! Нажмите /start чтобы вернуться в меню")
    admin := tb.User{73346375,"","","","",false}
    b.Send(
      &admin,
      fmt.Sprintf("Пользователь @%s забрал проект.", c.Sender.Username))
    return
  })

  b.Handle("/approve", func(m *tb.Message) {
    db.MustExec(`INSERT INTO SEworker(tid, approved) VALUES ($1, true)`, m.Payload)
    b.Send(m.Sender, "Новый полльзователь биржи добавлен.")
  })

  b.Handle("/workers", func(m *tb.Message) {
    workers := []SEworker{}
    db.Select(&workers, "SELECT * FROM SEworker")
    fmt.Println(workers)
    b.Send(m.Sender, "IDI NAHOOY")
  })

  b.Handle("/project", func(m *tb.Message) {
    err := client.Set(fmt.Sprintf("%d", m.Sender.ID), "project0", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(m.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }
    b.Send(m.Sender, "Введите информацию для проекта.")
  })

  b.Handle(&cancelProjectBtn, func(c *tb.Callback) {
    err := client.Set(fmt.Sprintf("%d", c.Sender.ID), "cancel0", 0).Err()
    if err != nil {
      log.Print(err)
      b.Send(c.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
      return
    }
    b.Send(c.Sender, "Очень жаль, что вы вынуждены отказаться от проекта. Пожалуйста, в следующем сообщении опишите максимально подробно причину отказа, что уже удалось сделать и на каком этапе находится проект. Чем больше информации будет предоставлено вами, тем быстрее будет принято решение по разрешению данного тикета. Спасибо!")
  })

  b.Handle(tb.OnText, func(m *tb.Message) {
    v, err := client.Get(fmt.Sprintf("%d", m.Sender.ID)).Result()
    if err != nil {
      log.Print(err)
    }
    position := fmt.Sprintf("%s", v)
    if err != nil {
      log.Print(err)
      b.Send(m.Sender, "К сожалению, что-то пошло не так. Пожалуйста, попробуйте позже. Администрация уже работает над решением проблемы! Чтобы вернуться в меню, нажмите /start")
      return
    }
    admin := tb.User{73346375,"","","","",false}
    switch position {
      case "qualify0":
        b.Send(&admin, fmt.Sprintf("%d – %s", m.Sender.ID, m.Sender.Username))
        b.Forward(&admin, m)
        b.Send(m.Sender, "Спасибо, администрация получила Вашу заявку и в самое ближайшее время свяжется с вами в Telegram! Вернитесь в меню с помощью /start")
        return
      case "askAdmin0":
        b.Forward(&admin, m)
        b.Send(m.Sender, "Спасибо, администрация получила Ваш вопрос и в самое ближайшее время свяжется с вами в Telegram! Вернитесь в меню с помощью /start")
        return
      case "techSupp0":
        b.Forward(&admin, m)
        b.Send(m.Sender, "Спасибо, администрация получила Ваш запрос тех. поддержки и в самое ближайшее время свяжется с вами в Telegram! Вернитесь в меню с помощью /start")
        return
      case "redeemMilestoneProject0":
        b.Forward(&admin, m)
        b.Send(m.Sender, "Спасибо, администрация получила Ваш запрос закрытие этапа/проекта и в самое ближайшее время свяжется с вами в Telegram! Вернитесь в меню с помощью /start")
        return
      case "project0":
        projectname = m.Text
        err := client.Set(fmt.Sprintf("%d", m.Sender.ID), "project1", 0).Err()
        if err != nil {
          log.Print(err)
          b.Send(m.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
          return
        }
        b.Send(m.Sender, "Описание для проекта:")
        return
      case "project1":
        projectdesc = m.Text
        err := client.Set(fmt.Sprintf("%d", m.Sender.ID), "project2", 0).Err()
        if err != nil {
          log.Print(err)
          b.Send(m.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
          return
        }
        b.Send(m.Sender, "Введите сложность проекта (1-5):")
      case "project2":
        projectdiff, err = strconv.Atoi(m.Text)
        if err != nil {
          b.Send(m.Sender, "Можно писать только цифры.")
          return
        }
        err := client.Set(fmt.Sprintf("%d", m.Sender.ID), "project3", 0).Err()
        if err != nil {
          log.Print(err)
          b.Send(m.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
          return
        }
        b.Send(m.Sender, "Цена проекта:")
      case "project3":
        projectprice, err = strconv.Atoi(m.Text)
        if err != nil {
          b.Send(m.Sender, "Можно писать только цифры.")
          return
        }
        tx := db.MustBegin()
        tx.MustExec(`INSERT INTO SEproject(name, description, difficulty, price, paid, progress, worker_id) VALUES ($1, $2, $3, $4, 0, 0, 0)`, projectname, projectdesc, projectdiff, projectprice)
        tx.Commit()
        b.Send(m.Sender, "Проект добавлен.")
        err := client.Set(fmt.Sprintf("%d", m.Sender.ID), "start", 0).Err()
        if err != nil {
          log.Print(err)
          b.Send(m.Sender, "Произошла ошибка, администрация уже получила запрос и работает на решением. Пожалуйста, воспользуйтесь сервисом позже")
          return
        }
        return
      case "cancel0":
        b.Send(&admin, fmt.Sprintf("Пользователь %d – @%s отказывается от проекта.", m.Sender.ID, m.Sender.Username))
        b.Forward(&admin, m)
        b.Send(m.Sender, "Спасибо, администрация получила Вашу заявку и в самое ближайшее время свяжется с вами в Telegram! Вернитесь в меню с помощью /start")
        return
      default:
        b.Send(m.Sender, "Я не понимаю обычный текст, нажмите /start")
    }
	})

	b.Start()
}
