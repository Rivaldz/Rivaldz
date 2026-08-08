package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Rivaldz/my-template-go/internal/controller/restapi/v1/request"
	"github.com/Rivaldz/my-template-go/internal/controller/restapi/v1/response"
	"github.com/Rivaldz/my-template-go/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Create post
// @Description Create a new post for the current user
// @ID          create-post
// @Tags        posts
// @Accept      json
// @Produce     json
// @Param       request body     request.CreatePost true "Post data"
// @Success     201     {object} entity.Post
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /posts [post]
func (r *V1) createPost(ctx *fiber.Ctx) error { //nolint:dupl // mirrors the standard create handler template
	userID, ok := authUserID(ctx)
	if !ok {
		return unauthorizedResponse(ctx)
	}

	var body request.CreatePost

	if !r.bindBody(ctx, "createPost", &body) {
		return badRequestResponse(ctx)
	}

	post, err := r.p.Create(ctx.UserContext(), userID, body.Title, body.Body)
	if err != nil {
		r.l.Error(err, "restapi - v1 - createPost")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(post)
}

// @Summary     List posts
// @Description List posts with pagination
// @ID          list-posts
// @Tags        posts
// @Produce     json
// @Param       limit  query    int    false "Limit"  default(10)
// @Param       offset query    int    false "Offset" default(0)
// @Success     200    {object} response.PostList
// @Failure     401    {object} response.Error
// @Failure     500    {object} response.Error
// @Security    BearerAuth
// @Router      /posts [get]
func (r *V1) listPosts(ctx *fiber.Ctx) error {
	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	posts, total, err := r.p.List(ctx.UserContext(), limit, offset)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listPosts")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.PostList{
		Posts: posts,
		Total: total,
	})
}

// @Summary     Get post
// @Description Get a post by ID
// @ID          get-post
// @Tags        posts
// @Produce     json
// @Param       id  path     string true "Post ID"
// @Success     200 {object} entity.Post
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /posts/{id} [get]
func (r *V1) getPost(ctx *fiber.Ctx) error {
	postID := ctx.Params("id")

	post, err := r.p.Get(ctx.UserContext(), postID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getPost")

		if errors.Is(err, entity.ErrPostNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "post not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(post)
}

// @Summary     Update post
// @Description Update a post (author only)
// @ID          update-post
// @Tags        posts
// @Accept      json
// @Produce     json
// @Param       id      path     string            true "Post ID"
// @Param       request body     request.UpdatePost  true "Updated post data"
// @Success     200     {object} entity.Post
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     403     {object} response.Error
// @Failure     404     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /posts/{id} [put]
func (r *V1) updatePost(ctx *fiber.Ctx) error { //nolint:dupl // mirrors the standard update handler template
	userID, ok := authUserID(ctx)
	if !ok {
		return unauthorizedResponse(ctx)
	}

	postID := ctx.Params("id")

	var body request.UpdatePost

	if !r.bindBody(ctx, "updatePost", &body) {
		return badRequestResponse(ctx)
	}

	post, err := r.p.Update(ctx.UserContext(), userID, postID, body.Title, body.Body)
	if err != nil {
		r.l.Error(err, "restapi - v1 - updatePost")

		if errors.Is(err, entity.ErrPostNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "post not found")
		}

		if errors.Is(err, entity.ErrPostForbidden) {
			return errorResponse(ctx, http.StatusForbidden, "forbidden")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(post)
}

// @Summary     Delete post
// @Description Delete a post and its comments (author only)
// @ID          delete-post
// @Tags        posts
// @Param       id  path     string true "Post ID"
// @Success     204 "No Content"
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /posts/{id} [delete]
func (r *V1) deletePost(ctx *fiber.Ctx) error { //nolint:dupl // mirrors the standard delete handler template
	userID, ok := authUserID(ctx)
	if !ok {
		return unauthorizedResponse(ctx)
	}

	err := r.p.Delete(ctx.UserContext(), userID, ctx.Params("id"))
	if err != nil {
		r.l.Error(err, "restapi - v1 - deletePost")

		if errors.Is(err, entity.ErrPostNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "post not found")
		}

		if errors.Is(err, entity.ErrPostForbidden) {
			return errorResponse(ctx, http.StatusForbidden, "forbidden")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}
