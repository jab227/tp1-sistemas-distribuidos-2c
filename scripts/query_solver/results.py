import numpy as np # linear algebra
import pandas as pd # data processing, CSV file I/O (e.g. pd.read_csv)
import os
import langid
import time
import sys

if len(sys.argv) != 3:
    print("Usage: results.py <games_file_name> <reviews_file_name>")
    sys.exit(1)

games_argv_position = 1
reviews_argv_position = 2
games_file_name = sys.argv[games_argv_position]
reviews_file_name = sys.argv[reviews_argv_position]

pd.set_option('display.max_columns', None)
pd.set_option('display.max_colwidth', 100)

games_df = pd.read_csv(f"{games_file_name}", header=None, skiprows=1)
reviews_df = pd.read_csv(f"{reviews_file_name}")

# Agrego columna "Unknown" ya que no viene nomenclada y genera un desfasaje en los indices de columnas
games_df_column_names = ['AppID', 'Name', 'Release date', 'Estimated owners', 'Peak CCU', 
                    'Required age', 'Price', 'Unknown', 'DiscountDLC count', 'About the game', 
                    'Supported languages', 'Full audio languages', 'Reviews', 'Header image', 
                    'Website', 'Support url', 'Support email', 'Windows', 'Mac', 
                    'Linux', 'Metacritic score', 'Metacritic url', 'User score', 
                    'Positive', 'Negative', 'Score rank', 'Achievements', 
                    'Recommendations', 'Notes', 'Average playtime forever', 
                    'Average playtime two weeks', 'Median playtime forever', 
                    'Median playtime two weeks', 'Developers', 'Publishers', 
                    'Categories', 'Genres', 'Tags', 'Screenshots', 'Movies']
games_df.columns = games_df_column_names

# Dataframes cleaning
games_df_columns = ['AppID', 'Name', 'Windows', 'Mac', 'Linux', 'Genres', 'Release date', 'Average playtime forever', 'Positive', 'Negative']
reviews_df_columns = ['app_id', 'review_text', 'review_score']
games_df_cleaned = games_df.dropna(subset=games_df_columns)[games_df_columns].copy()
reviews_df_cleaned = reviews_df.dropna(subset=reviews_df_columns)[reviews_df_columns].copy()

games_df_cleaned["Genres"] = games_df_cleaned["Genres"].str.lower()
reviews_df_cleaned['review_text'] = reviews_df_cleaned['review_text'].astype(str)

# Queries

# Query 1
# Cantidad de juegos soportados en cada plataforma (Windows, Linux, MAC)
windows_supported_games = games_df_cleaned[games_df_cleaned["Windows"] == True]
linux_supported_games = games_df_cleaned[games_df_cleaned["Linux"] == True]
mac_supported_games = games_df_cleaned[games_df_cleaned["Mac"] == True]

print("===========")
print("Query 1:")
print("===========")
print(f"windows: {str(windows_supported_games.shape[0])}")
print(f"mac: {str(mac_supported_games.shape[0])}")
print(f"linux: {str(linux_supported_games.shape[0])}")

# Query 2
# Nombre de los juegos top 10 del género "Indie" publicados en la década del 2010 con más tiempo promedio histórico de juego
games_indie = games_df_cleaned[games_df_cleaned["Genres"].str.contains("indie")]
games_indie_2010_decade = games_indie[games_indie["Release date"].str.contains("201")]
q2_result = games_indie_2010_decade.sort_values(by='Average playtime forever', ascending=False).head(10)

print("===========")
print("Query 2:")
print("===========")
for index, value in enumerate(q2_result[['Name', 'Average playtime forever']]['Name'], start=1):
    print(f"{index}: {value}")

# Query 3
# Nombre de los juegos top 5 del género "Indie" con más reseñas positivas
games_indie_reduced = games_indie[["AppID", "Name"]]
reviews_reduced_q3 = reviews_df_cleaned[["app_id", "review_score"]]
games_indie_reviews = pd.merge(games_indie_reduced, reviews_reduced_q3, left_on='AppID', right_on='app_id', how='inner')

def positive_score(score):
    return 1 if score > 0 else 0

games_indie_reviews['positive_score'] = games_indie_reviews['review_score'].apply(positive_score)
q3_result = games_indie_reviews.groupby('Name')['positive_score'].sum().sort_values(ascending=False).head(5)

print("===========")
print("Query 3:")
print("===========")
for index, value in enumerate(q3_result.head(10).index, start=1):
    print(f"{index}: {value}")

# Query 4
# Nombre de juegos del género "action" con más de 5.000 reseñas negativas en idioma inglés
games_action = games_df_cleaned[games_df_cleaned["Genres"].str.contains("action")]
games_action_reduced = games_action[["AppID", "Name"]]
reviews_q4 = reviews_df_cleaned.copy()

def negative_score(score):
    return 1 if score < 0 else 0

reviews_q4["negative_score"] = reviews_q4["review_score"].apply(negative_score)
reviews_q4_negatives = reviews_q4[reviews_q4["negative_score"] == 1].copy()
reviews_count = reviews_q4_negatives.groupby('app_id').size().reset_index(name='count')
reviews_count_more_than_5000 = reviews_count[reviews_count["count"] > 5000]

games_action_with_5000_negative_reviews = pd.merge(games_action_reduced, reviews_count_more_than_5000, left_on='AppID', right_on="app_id", how='inner')
games_action_with_5000_negative_reviews = games_action_with_5000_negative_reviews[["AppID", "Name"]]

reviews_count_more_than_5000_with_text = pd.merge(reviews_q4, games_action_with_5000_negative_reviews, left_on='app_id', right_on="AppID", how='inner')
reviews_count_more_than_5000_with_text = reviews_count_more_than_5000_with_text[["app_id", "review_text"]]

def detect_language(texto):
    language, _ = langid.classify(texto)
    return language

reviews_count_more_than_5000_with_text["review_language"] = reviews_count_more_than_5000_with_text['review_text'].apply(detect_language)
reviews_count_more_than_5000_with_text_english = reviews_count_more_than_5000_with_text[reviews_count_more_than_5000_with_text["review_language"] == "en"]

q4_results_app_ids = reviews_count_more_than_5000_with_text_english.groupby('app_id').size().reset_index(name='count')
q4_results_app_ids = q4_results_app_ids[q4_results_app_ids["count"] > 5000]
q4_results_games_names = pd.merge(q4_results_app_ids, games_action_with_5000_negative_reviews, left_on='app_id', right_on="AppID", how='inner')["Name"]

print("===========")
print("Query 4:")
print("===========")
for index, value in enumerate(q4_results_games_names.head(25).sort_values(), start=1):
    print(f"{index}: {value}")

# Query 5
# Nombre de juegos del género "action" dentro del percentil 90 en cantidad de reseñas negativas
reviews_q5 = reviews_df_cleaned.copy()
reviews_q5 = reviews_q5[["app_id", "review_score"]]
reviews_q5["negative_score"] = reviews_q5["review_score"].apply(negative_score)
reviews_q5_negative_score = reviews_q5[reviews_q5["negative_score"] == 1]
reviews_q5_negative_score_action = pd.merge(reviews_q5_negative_score, games_action_reduced, left_on='app_id', right_on="AppID", how='inner')
reviews_q5_negative_score_action_by_app_id = reviews_q5_negative_score_action.groupby('app_id').size().reset_index(name='count')
percentil_90 = reviews_q5_negative_score_action_by_app_id['count'].quantile(0.90)
q5_result = reviews_q5_negative_score_action_by_app_id[reviews_q5_negative_score_action_by_app_id['count'] >= percentil_90]
q5_result_with_game_names = pd.merge(q5_result, games_action_reduced, left_on='app_id', right_on="AppID", how='inner')

print("===========")
print("Query 5:")
print("===========")
for index, value in enumerate(q5_result_with_game_names[["app_id", "Name"]]['Name'].head(10).sort_values(), start=1):
    print(f"{index}: {value}")

