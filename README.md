# Checklist do Projeto: Servidor de Chat em Tempo Real

## Fase 0: Preparação e Planejamento

[ ] 1. Setup do Ambiente: Instalar o Go e configurar o espaço de trabalho.

[ ] 2. Estrutura do Projeto: Definir a organização das pastas e arquivos.

[ ] 3. Escolha do Banco de Dados: Decidir entre PostgreSQL, MongoDB ou Redis e entender o porquê da escolha.

[ ] 4. Definição do "Contrato": Desenhar como serão as mensagens trocadas via WebSocket (a estrutura JSON).

## Fase 1: O Básico Funcional (O Coração do Chat)

[ ] 1. Servidor HTTP Básico: Criar um servidor web simples em Go.

[ ] 2. Endpoint WebSocket: Criar uma rota que "transforma" uma conexão HTTP em uma conexão WebSocket.

[ ] 3. O Hub de Conexões: Implementar uma estrutura central para gerenciar todos os usuários conectados (registrar, remover e enviar mensagens para todos).

[ ] 4. Broadcast de Mensagens: Fazer o Hub enviar as mensagens recebidas para todos os clientes conectados.

## Fase 2: Autenticação e Usuários

[ ] 1. Conexão com o PostgreSQL: Configurar e testar a conexão do nosso app Go com o banco.

[ ] 2. Modelagem e Criação da Tabela: Criar a tabela users no banco e o struct correspondente em Go.

[ ] 3. Endpoint de Registro (/register): Permitir que novos usuários se cadastrem de forma segura (com hash de senha).

[ ] 4. Endpoint de Login (/login): Validar as credenciais e, se corretas, gerar e retornar um JWT.

[ ] 5. Proteção do WebSocket: Modificar a rota /ws para que ela só aceite conexões que apresentem um JWT válido.

## Fase 3: Lógica do Chat e Persistência de Dados

[ ] 1. Modelagem do Banco: Desenhar as tabelas/coleções para users, rooms e messages.

[ ] 2. Lógica de Salas: Permitir que usuários criem, entrem e saiam de salas de chat. O Hub precisa saber quem está em qual sala.

[ ] 3. Mensagens Privadas (DMs): Implementar a lógica para enviar uma mensagem de um usuário específico para outro.

[ ] 4. Persistência de Mensagens: Salvar cada mensagem enviada no banco de dados.

[ ] 5. Carregamento de Histórico: Ao entrar em uma sala, carregar as últimas X mensagens do banco.

## Fase 4: Robustez e Deploy (Deixando Profissional)

[ ] 1. Lógica de Reconexão: Garantir que o servidor lide bem com clientes que caem e voltam.

[ ] 2. Dockerização: Criar um Dockerfile para empacotar nossa aplicação em um contêiner.

[ ] 3. Deploy Simples: Usar Docker Compose para rodar a aplicação e o banco de dados juntos.

[ ] 4. Monitoramento (Bônus): Configurar métricas básicas com Prometheus para ficar de olho na saúde do servidor.
