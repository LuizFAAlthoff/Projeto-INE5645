import multiprocessing
import threading
import time
import random

# -------------------------
# Fila manual com locks
# -------------------------
class FilaProcesso:
    def __init__(self, manager):
        self.itens = manager.list()
        self.lock = multiprocessing.Lock()
        self.cond = multiprocessing.Condition(self.lock)

    def put(self, item):
        with self.cond:
            self.itens.append(item)
            self.cond.notify()

    def get(self):
        with self.cond:
            while len(self.itens) == 0:
                self.cond.wait()
            return self.itens.pop(0)

# -------------------------
# Future manual
# -------------------------
class FutureCaseiro:
    def __init__(self, manager):
        self.resultado = manager.dict()
        self.resultado['concluido'] = False
        self.cond = threading.Condition()

    def set_resultado(self, valor):
        with self.cond:
            self.resultado['valor'] = valor
            self.resultado['concluido'] = True
            self.cond.notify_all()

    def result(self):
        with self.cond:
            while not self.resultado['concluido']:
                self.cond.wait()
            return self.resultado['valor']

# -------------------------
# Função de transformação/carregamento
# -------------------------
def transformar_carregar(item):
    print(f"[Worker] Transformando {item}")
    time.sleep(random.uniform(0.2, 0.4))
    resultado = item * 10
    print(f"[Worker] Carregado: {resultado}")
    return resultado

# -------------------------
# ThreadPool caseiro
# -------------------------
class ThreadPool:
    def __init__(self, num_threads):
        self.tasks = []
        self.lock = threading.Lock()
        self.cond = threading.Condition(self.lock)
        self.workers = []
        self.parar = False
        for _ in range(num_threads):
            t = threading.Thread(target=self.worker)
            t.start()
            self.workers.append(t)

    def worker(self):
        while True:
            with self.cond:
                while not self.tasks and not self.parar:
                    self.cond.wait()
                if self.parar and not self.tasks:
                    break
                func, arg, future = self.tasks.pop(0)
            try:
                resultado = func(arg)
                future.set_resultado(resultado)
            except Exception as e:
                future.set_resultado(f"[Erro]: {e}")

    def submit(self, func, arg, future):
        with self.cond:
            self.tasks.append((func, arg, future))
            self.cond.notify()

    def shutdown(self):
        with self.cond:
            self.parar = True
            self.cond.notify_all()
        for t in self.workers:
            t.join()

# -------------------------
# Produtor
# -------------------------
def produtor(dados, fila):
    for item in dados:
        print(f"[Produtor] Extraindo: {item}")
        fila.put(item)
        time.sleep(random.uniform(0.1, 0.2))
    fila.put(None)  # SENTINEL

# -------------------------
# Consumidor (com ThreadPool interno)
# -------------------------
def consumidor(fila, futuros, manager):
    pool = ThreadPool(num_threads=4)

    while True:
        item = fila.get()
        if item is None:
            break
        future = FutureCaseiro(manager)
        pool.submit(transformar_carregar, item, future)
        futuros.append(future)

    pool.shutdown()

# -------------------------
# Programa principal
# -------------------------
def main():
    dados = list(range(1, 11))
    manager = multiprocessing.Manager()
    fila = FilaProcesso(manager)
    futuros = manager.list()

    p_produtor = multiprocessing.Process(target=produtor, args=(dados, fila))
    p_consumidor = multiprocessing.Process(target=consumidor, args=(fila, futuros, manager))

    p_produtor.start()
    p_consumidor.start()

    p_produtor.join()
    p_consumidor.join()

    print("\n[Main] Aguardando resultados...")
    for future in list(futuros):
        print("[Main] Resultado final:", future.result())

    print("\n[Main] ETL concluído com sucesso.")

if __name__ == "__main__":
    multiprocessing.set_start_method("fork")  # use 'spawn' no Windows
    main()
