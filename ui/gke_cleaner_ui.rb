require 'rubygems'
require 'sinatra'
require 'net/http'
require 'json'

class GKECleanerUi < Sinatra::Base
  set :basic_auth_username, ENV.fetch('BASIC_AUTH_USERNAME')
  set :basic_auth_password, ENV.fetch('BASIC_AUTH_PASSWORD')
  set :gke_cleaner_backend_url, ENV.fetch('GKE_CLEANER_BACKEND_URL')

  use Rack::Auth::Basic, "Restricted" do |username, password|
    username == settings.basic_auth_username && password == settings.basic_auth_password
  end

  get '/' do
    uri = URI("#{settings.gke_cleaner_backend_url}/clusters")

    req = Net::HTTP::Get.new(uri)
    req.basic_auth settings.basic_auth_username, settings.basic_auth_password

    res = Net::HTTP.start(uri.hostname, uri.port, :use_ssl => true) {|http|
      http.request(req)
    }

    if res.is_a?(Net::HTTPSuccess)
      @clusters = JSON.parse(res.body)
      erb :index
    else
      status 500
    end
  end

  post '/renew/:name' do
    uri = URI("#{settings.gke_cleaner_backend_url}/clusters/renew/#{params['name']}")
    req = Net::HTTP::Post.new(uri)
    req.basic_auth settings.basic_auth_username, settings.basic_auth_password

    res = Net::HTTP.start(uri.hostname, uri.port, :use_ssl => true) {|http|
      http.request(req)
    }

    if res.is_a?(Net::HTTPSuccess)
      redirect '/'
    else
      status 500
    end
  end

  post '/ignore/:name' do
    uri = URI("#{settings.gke_cleaner_backend_url}/clusters/ignore/#{params['name']}")
    req = Net::HTTP::Post.new(uri)
    req.basic_auth settings.basic_auth_username, settings.basic_auth_password

    res = Net::HTTP.start(uri.hostname, uri.port, :use_ssl => true) {|http|
      http.request(req)
    }

    if res.is_a?(Net::HTTPSuccess)
      redirect '/'
    else
      status 500
    end
  end

  post '/unignore/:name' do
    uri = URI("#{settings.gke_cleaner_backend_url}/clusters/unignore/#{params['name']}")
    req = Net::HTTP::Post.new(uri)
    req.basic_auth settings.basic_auth_username, settings.basic_auth_password

    res = Net::HTTP.start(uri.hostname, uri.port, :use_ssl => true) {|http|
      http.request(req)
    }

    if res.is_a?(Net::HTTPSuccess)
      redirect '/'
    else
      status 500
    end
  end
end
