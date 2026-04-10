enable_testing()
add_executable(test_runner test_runner.cpp)
target_link_libraries(test_runner PRIVATE common llama llama_core)
add_test(NAME RunnerTest COMMAND test_runner)

add_executable(test_runner_gen test_runner_gen.cpp)
target_link_libraries(test_runner_gen PRIVATE common llama llama_core)
add_test(NAME RunnerTestGen COMMAND test_runner_gen)

add_executable(test_runner_chat test_runner_chat.cpp)
target_link_libraries(test_runner_chat PRIVATE common llama llama_core)
add_test(NAME RunnerTestChat COMMAND test_runner_chat)

add_executable(test_embedding test_embedding.cpp)
target_link_libraries(test_embedding PRIVATE common llama llama_core)
add_test(NAME EmbeddingTest COMMAND test_embedding)